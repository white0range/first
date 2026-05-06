package controllers

import (
	"context"
	"encoding/json" // 记得引入 JSON 包
	"gojo/global"
	"gojo/models"
	"gojo/utils"
	"io"
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
		"user_id":       userID.(uint),
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
		"code":          submission.Code,
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
	////////////////////////////

	var submissions []models.Submission
	var total int64

	// ==========================================
	// 🔍 阶段一：拼装基础查询条件 (不要立刻 Find)
	// ==========================================
	// 提取出一个公共的 query 对象，死死锁定当前用户的 ID
	query := models.DB.Model(&models.Submission{}).Where("user_id = ?", userID)

	// ==========================================
	// 📊 阶段二：查总数 (干干净净地查，没有 Limit 和 Offset)
	// ==========================================
	query.Count(&total)

	// ==========================================
	// 💥 阶段三：挂载分页神器 & 排序剔除 & 获取数据
	// ==========================================
	// 直接在刚才的 query 上调用 .Scopes(utils.Paginate(c))
	if err := query.Scopes(utils.Paginate(c)).
		Order("created_at desc").      // 最新提交排前面
		Omit("code", "actual_output"). // 剔除大体积字段，保护带宽
		Find(&submissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，获取记录失败"})
		return
	}

	// 补充：如果你前端非要在返回值里要 page 和 limit 回显，简单提取一下即可
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

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

// GetAIAssistance 呼叫 AI 导师分析某次提交
func GetAIAssistance(c *gin.Context) {
	submissionID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "系统异常：无法获取当前用户身份"})
		return
	}

	// 1. 查出这条战绩
	var submission models.Submission
	if err := models.DB.First(&submission, submissionID).Error; err != nil {
		c.JSON(404, gin.H{"error": "找不到该战绩"})
		return
	}

	// 2. 权限校验：只能让 AI 分析自己的代码（防白嫖和防泄露）
	if submission.UserID != userID {
		c.JSON(403, gin.H{"error": "您只能呼叫 AI 诊断自己的代码！"})
		return
	}

	// 3. 如果是 AC (通过) 的代码，导师不屑于看
	if submission.Status == "AC" {
		c.JSON(400, gin.H{"message": "代码已经完美通过，无需导师诊断！"})
		return
	}

	// 2. 极其关键：设置响应头，告诉前端“我要开始喷水了，别挂断！”
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// 3. 💥 优雅调用 utils 里的底层能力，拿到“水龙头”
	stream, err := utils.AskAIStream(submission.Code, submission.Language, submission.ActualOutput)
	if err != nil {
		c.SSEvent("error", "AI 导师暂时无法连接")
		return
	}
	defer stream.Close() // 离开这个接口时，务必关掉底层的大模型连接！
	// 4. 召唤 AI (这一步可能会卡顿 2-5 秒)
	//aiAdvice, err := utils.AskAI(submission.Code, submission.Language, submission.ActualOutput)
	//if err != nil {
	//	c.JSON(500, gin.H{"error": "AI 导师正在休息，请稍后再试"})
	//	return
	//}
	//// 5. 将导师的建议返回给前端
	//c.JSON(200, gin.H{
	//	"message": "AI 诊断完成",
	//	"advice":  aiAdvice,
	//})

	// 4. 开始把接到的水，转发给前端
	c.Stream(func(w io.Writer) bool {
		// 从 utils 的水龙头里接一滴水
		response, err := stream.Recv()

		if err != nil {
			// 如果遇到 EOF (End Of File)，说明 AI 说完了，正常结束连接
			// 如果是其他 err，说明中途断了，也得结束连接
			return false
		}

		// 把大模型吐出来的这个字，立刻推给前端
		c.SSEvent("message", response.Choices[0].Delta.Content)

		return true // 告诉 Gin：我还没发完，继续保持循环！
	})
}
