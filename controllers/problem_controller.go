package controllers

import (
	"gojo/models"
	"gojo/services"
	"net/http"
	"strconv" //是 Go 标准库里做字符串和基本类型转换的包

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateProblem 接收前端发来的题目数据，并存入数据库
func CreateProblem(c *gin.Context) {
	// 1. 拿出咱们的“接客表单”（DTO）
	var req models.ProblemRequest

	// 2. 门卫查验：解析 JSON 并触发校验
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误，请检查必填项：标题(title)和描述(description)"})
		return
	}

	// 3. 呼叫 Service 大脑干活！(传入 c.Request.Context() 以便 Redis 和底层链路追踪使用)
	problem, err := services.CreateProblem(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，题目发布失败"})
		return
	}

	// 4. 喜报！成功返回
	c.JSON(http.StatusOK, gin.H{
		"message":    "题目发布成功！",
		"problem_id": problem.ID, // Service 返回的对象里已经有了最新的主键 ID
	})
}

// GetProblemList 获取题目列表（菜单）
func GetProblemList(c *gin.Context) {
	// 1. 从 URL 的问号后面拿分页参数，如果没传，就给默认值
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	tagIDStr := c.Query("tag_id") // 👈 新增：尝试获取标签 ID，如果没传就是空字符串 ""
	// 转换成纯数字
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	// 🛡️ 极其关键的安全处理：优雅地获取 userID，防止断言宕机
	var uid uint = 0
	if id, exists := c.Get("userid"); exists {
		uid = id.(uint) // 只有确保存在时，才进行类型断言
	}

	// 呼叫大脑：把纯粹的数字和字符串扔给 Service
	res, err := services.GetProblemList(c.Request.Context(), page, limit, tagIDStr, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取列表失败"})
		return
	}

	// 正常返回
	c.JSON(http.StatusOK, res)
}

// GetProblemDetail 获取单个题目详情（菜品详情）
func GetProblemDetail(c *gin.Context) {
	id := c.Param("id")

	// 呼叫 Service 干活
	problem, err := services.GetProblemDetail(c.Request.Context(), id)

	if err != nil {
		// 拦截：返回 404 和错误信息，并极其关键地 return 切断流程！
		c.JSON(http.StatusNotFound, gin.H{
			"message": "这道题不存在或已被删除！",
			"data":    nil,
		})
		return
	}

	// 成功：返回 200 和数据
	c.JSON(http.StatusOK, gin.H{
		"message": "获取题目详情成功",
		"data":    problem,
	})
}

// controllers/problem_controller.go

// UpdateProblem 修改题目（管理员专属）
func UpdateProblem(c *gin.Context) {
	// 1. 获取要修改的题目 ID
	problemID := c.Param("id")

	// 2. 拿出“接客表单”
	var req models.ProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误，请检查传入的数据格式"})
		return
	}

	// 3. 呼叫 Service 大脑干活！(传入 c.Request.Context() 贯穿链路)
	if err := services.UpdateProblem(c.Request.Context(), problemID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，题目更新失败"})
		return
	}

	// 4. 返回喜报
	c.JSON(http.StatusOK, gin.H{
		"message": "题目修改成功，天下大吉！",
	})
}

// controllers/problem_controller.go

// DeleteProblem 删除题目（管理员专属）
func DeleteProblem(c *gin.Context) {
	// 1. 获取 ID
	problemID := c.Param("id")

	// 2. 呼叫 Service 执行“处决”
	if err := services.DeleteProblem(c.Request.Context(), problemID); err != nil {
		// 如果是“记录未找到”，返回 404
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "斩杀失败：这道题根本不存在！"})
			return
		}
		// 其他错误返回 500
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，题目删除失败"})
		return
	}

	// 3. 返回成功
	c.JSON(http.StatusOK, gin.H{
		"message": "题目已成功删除",
	})
}

// controllers/problem_controller.go

// UpdateProblemTagsRequest 接收前端传来的全新标签 ID 数组 (建议移到 models 目录)
type UpdateProblemTagsRequest struct {
	TagIDs []uint `json:"tag_ids"`
}

// UpdateProblemTags 更新题目的标签（全量替换）
func UpdateProblemTags(c *gin.Context) {
	problemID := c.Param("id")

	var req UpdateProblemTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	// 呼叫 Service 执行业务
	if err := services.UpdateProblemTags(c.Request.Context(), problemID, req.TagIDs); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新标签关联失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "题目标签更新成功"})
}
