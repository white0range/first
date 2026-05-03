package controllers

import (
	"context"
	"encoding/json" // 记得引入 JSON 包
	"gojo/global"
	"gojo/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SubmitCode 接收玩家提交的代码
func SubmitCode(c *gin.Context) {
	// 1. 拿出提交表单
	var req models.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：缺少题目ID、语言或代码"})
		return
	}

	// 2. 极其核心的一步：从保安（中间件）那里拿到当前玩家的 ID
	// 绝不能让前端传 UserID，否则黑客可以伪造别人的身份交代码！
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "系统异常：无法获取当前用户身份"})
		return
	}

	// 3. 把收到的代码组装成要下硬盘的实体对象
	submission := models.Submission{
		// 注意：中间件里存的 userID 默认是 interface{}，需要断言回 uint
		UserID:    userID.(uint),
		ProblemID: req.ProblemID,
		Language:  req.Language,
		Code:      req.Code,
		// Status 不需要写，GORM 会自动触发 default:'Pending'
	}

	// 4. 落库保存！
	if err := models.DB.Create(&submission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，代码提交失败"})
		return
	}

	// ==========================================
	// 🚀 5. 【极其核心的改造】：不再开启协程，而是扔进 Redis 队列！
	// ==========================================

	// 5.1 把工人干活需要的所有参数，打包成一个结构体
	task := map[string]interface{}{
		"submission_id": submission.ID,
		"problem_id":    req.ProblemID,
		"code":          req.Code,
	}

	// 5.2 序列化：Redis 队列只认识字符串，所以我们要把它变成 JSON 字符串
	taskBytes, err := json.Marshal(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常：任务序列化失败"})
		return
	}

	// 5.3 发货！把 JSON 字符串狠狠地塞进名为 "judge_queue" 的 Redis 传送带左侧
	// Rdb.LPush 会返回塞入之后队列里的任务总数，但我们这里不需要管它，直接处理报错即可
	err = global.Rdb.LPush(context.Background(), "judge_queue", taskBytes).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常：无法将任务加入判题队列"})
		return
	}

	// 6. 告诉玩家：大堂经理已经把单子挂在厨房了，你可以回去等吃了！
	c.JSON(http.StatusOK, gin.H{
		"message":       "代码已成功加入判题队列！",
		"submission_id": submission.ID,
		"status":        "Pending",
	})

}

// GetSubmissionResult 获取某次代码提交的最终战果（查快递接口）
func GetSubmissionResult(c *gin.Context) {
	// 1. 从 URL 路径中拿到物流单号 (例如 /api/submissions/5 中的 "5")
	id := c.Param("id")

	// 2. 去数据库里查这条记录
	var submission models.Submission
	// First 方法会自动生成类似 SELECT * FROM submissions WHERE id = 5 的 SQL
	if err := models.DB.First(&submission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "客官，找不到这条提交记录！"})
		return
	}

	// 3. 将当前的最新状态返回给前端
	c.JSON(http.StatusOK, gin.H{
		"message":       "获取状态成功",
		"submission_id": submission.ID,
		"problem_id":    submission.ProblemID,
		"status":        submission.Status,       // 可能是 Pending, AC, WA, CE, SE 等
		"actual_output": submission.ActualOutput, // 如果报错了，这里有详细日志
		"language":      submission.Language,
	})
}

// GetMySubmissions 获取当前玩家的提交历史（带分页）
func GetMySubmissions(c *gin.Context) {
	// 1. 极其关键的身份校验：只能查自己的！
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "系统异常：无法获取当前用户身份"})
		return
	}

	// 2. 提取分页参数 (和大厅列表一样)
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	var submissions []models.Submission
	var total int64

	// 3. 递空名片，查总数 (注意：要加上 Where 条件，只查当前用户的！)
	models.DB.Model(&models.Submission{}).
		Where("user_id = ?", userID).
		Count(&total)

	// 4. 查具体的战绩列表
	// 【大厂细节 1】：Order("created_at desc") 按照时间倒序，最新提交的排在最前面！
	// 【大厂细节 2】：Omit("code", "actual_output") 如果不需要在列表页展示几百行的代码和长篇报错日志，可以用 Omit 剔除这些大字段，极大地节省网络带宽。
	models.DB.Where("user_id = ?", userID).
		Order("created_at desc").      // desc 是 descending (降序) 的缩写
		Omit("code", "actual_output"). // 列表页不查大文本（可选项，看你的前端设计）
		Limit(limit).
		Offset(offset).
		Find(&submissions)

	// 5. 组装并返回
	c.JSON(http.StatusOK, gin.H{
		"message": "获取提交历史成功",
		"data": gin.H{
			"items": submissions,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}
