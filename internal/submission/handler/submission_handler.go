package handler

import (
	"gojo/internal/submission/dto"

	"gojo/internal/submission/service"

	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SubmissionHandler struct {
	svc *service.SubmissionService
}

func NewSubmissionHandler(svc *service.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{svc: svc}
}

// SubmitCode 接收玩家提交的代码
func (h *SubmissionHandler) SubmitCode(c *gin.Context) {
	var req dto.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：缺少题目ID、语言或代码"})
		return
	}

	// 🛡️ 极其核心且安全的获取身份方式
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "系统异常：无法获取当前用户身份"})
		return
	}

	// 💣 防治宕机：如果前置中间件放进来的不是 uint（比如 JWT 解析出的是 float64），这里会 Panic！
	// 加上 ok 校验，这是 Go 语言极其重要的防御性编程习惯！
	userID, ok := userIDRaw.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常：用户身份解析失败"})
		return
	}

	// 呼叫 Service 干活！
	submission, err := h.svc.SubmitCode(c.Request.Context(), userID, req)
	if err != nil {
		// 这里可以打印 err 到日志里，方便排查
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，代码提交失败"})
		return
	}

	// 告诉玩家：大堂经理已经把单子挂在厨房了，回去等吃吧！
	c.JSON(http.StatusOK, gin.H{
		"message":       "代码已成功加入判题队列！",
		"submission_id": submission.ID,
		"status":        "Pending",
	})
}

// controllers/submission_controller.go

// GetSubmissionResult 获取某次代码提交的最终战果（查快递接口）
func (h *SubmissionHandler) GetSubmissionResult(c *gin.Context) {
	// 1. 拿快递单号
	id := c.Param("id")

	// 2. 🛡️ 极其核心且安全的获取当前“取件人”的身份
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录后再查询战果！"})
		return
	}
	currentUserID, ok := userIDRaw.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户身份解析失败"})
		return
	}

	// 3. 呼叫 Service 大脑查快递！
	submission, err := h.svc.GetSubmissionResult(c.Request.Context(), id, currentUserID)
	if err != nil {
		// 判断是查不到，还是没权限查
		if err.Error() == "unauthorized access" {
			c.JSON(http.StatusForbidden, gin.H{"error": "警告：你没有权限查看别人的代码！"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "客官，找不到这条提交记录！"})
		return
	}

	// 4. 返回战报
	c.JSON(http.StatusOK, gin.H{
		"message":       "获取状态成功",
		"submission_id": submission.ID,
		"problem_id":    submission.ProblemID,
		"status":        submission.Status,
		"actual_output": submission.ActualOutput,
		"code":          submission.Code,
		"language":      submission.Language,
	})
}

// controllers/submission_controller.go

// GetMySubmissions 获取当前玩家的提交历史（带分页）
func (h *SubmissionHandler) GetMySubmissions(c *gin.Context) {
	// 1. 🛡️ 极其关键的安全校验与断言
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
		return
	}
	userID, ok := userIDRaw.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户身份解析异常"})
		return
	}

	// 2. 提取分页参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	// 3. 呼叫 Service 大脑
	res, err := h.svc.GetMySubmissions(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，获取记录失败"})
		return
	}

	// 4. 组装并返回标准格式
	c.JSON(http.StatusOK, gin.H{
		"message": "获取提交历史成功",
		"data":    res,
	})
}

//// GetAIAssistance 呼叫 AI 导师分析某次提交
//func GetAIAssistance(c *gin.Context) {
//	submissionID := c.Param("id")
//	userID, exists := c.Get("userID")
//	if !exists {
//		c.JSON(http.StatusUnauthorized, gin.H{"error": "系统异常：无法获取当前用户身份"})
//		return
//	}
//
//	// 1. 查出这条战绩
//	var submission models.Submission
//	if err := models.DB.First(&submission, submissionID).Error; err != nil {
//		c.JSON(404, gin.H{"error": "找不到该战绩"})
//		return
//	}
//
//	// 2. 权限校验：只能让 AI 分析自己的代码（防白嫖和防泄露）
//	if submission.UserID != userID {
//		c.JSON(403, gin.H{"error": "您只能呼叫 AI 诊断自己的代码！"})
//		return
//	}
//
//	// 3. 如果是 AC (通过) 的代码，导师不屑于看
//	if submission.Status == "AC" {
//		c.JSON(400, gin.H{"message": "代码已经完美通过，无需导师诊断！"})
//		return
//	}
//
//	// 2. 极其关键：设置响应头，告诉前端“我要开始喷水了，别挂断！”
//	c.Writer.Header().Set("Content-Type", "text/event-stream")
//	c.Writer.Header().Set("Cache-Control", "no-cache")
//	c.Writer.Header().Set("Connection", "keep-alive")
//
//	// 3. 💥 优雅调用 utils 里的底层能力，拿到“水龙头”
//	stream, err := utils.AskAIStream(submission.Code, submission.Language, submission.ActualOutput)
//	if err != nil {
//		c.SSEvent("error", "AI 导师暂时无法连接")
//		return
//	}
//	defer stream.Close() // 离开这个接口时，务必关掉底层的大模型连接！
//	// 4. 召唤 AI (这一步可能会卡顿 2-5 秒)
//	//aiAdvice, err := utils.AskAI(submission.Code, submission.Language, submission.ActualOutput)
//	//if err != nil {
//	//	c.JSON(500, gin.H{"error": "AI 导师正在休息，请稍后再试"})
//	//	return
//	//}
//	//// 5. 将导师的建议返回给前端
//	//c.JSON(200, gin.H{
//	//	"message": "AI 诊断完成",
//	//	"advice":  aiAdvice,
//	//})
//
//	// 4. 开始把接到的水，转发给前端
//	c.Stream(func(w io.Writer) bool {
//		// 从 utils 的水龙头里接一滴水
//		response, err := stream.Recv()
//
//		if err != nil {
//			// 如果遇到 EOF (End Of File)，说明 AI 说完了，正常结束连接
//			// 如果是其他 err，说明中途断了，也得结束连接
//			return false
//		}
//
//		// 把大模型吐出来的这个字，立刻推给前端
//		c.SSEvent("message", response.Choices[0].Delta.Content)
//
//		return true // 告诉 Gin：我还没发完，继续保持循环！
//	})
//}

// controllers/ai_controller.go

// GetAIAssistance 呼叫 AI 导师分析某次提交 (SSE 流式响应)
func (h *SubmissionHandler) GetAIAssistance(c *gin.Context) {
	submissionID := c.Param("id")

	// 1. 🛡️ 安全提取身份
	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "系统异常：无法获取当前用户身份"})
		return
	}
	userID, ok := userIDRaw.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户身份解析失败"})
		return
	}

	// 2. 呼叫 Service 大脑，索要水龙头
	stream, err := h.svc.GetAIAssistanceStream(c.Request.Context(), submissionID, userID)
	if err != nil {
		// 根据 Service 返回的错误类型，进行精准的 HTTP 状态码拦截
		switch err.Error() {
		case "not_found":
			c.JSON(http.StatusNotFound, gin.H{"error": "找不到该战绩"})
		case "forbidden":
			c.JSON(http.StatusForbidden, gin.H{"error": "您只能呼叫 AI 诊断自己的代码！"})
		case "already_ac":
			c.JSON(http.StatusBadRequest, gin.H{"message": "代码已经完美通过，无需导师诊断！"})
		case "ai_connect_failed":
			c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 导师暂时无法连接"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "系统未知异常"})
		}
		return
	}

	// 🚨 极其关键：离开这个接口时，务必关掉底层的大模型连接！
	defer stream.Close()

	// 3. 设置响应头，告诉前端“我要开始喷水了，别挂断！”
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// 4. 开始把接到的水，转发给前端
	c.Stream(func(w io.Writer) bool {
		// 从 Service 拿到的水龙头里接一滴水
		response, err := stream.Recv()

		if err != nil {
			// 如果遇到 EOF (End Of File)，说明 AI 说完了，正常结束连接
			// 如果是其他 err (比如用户强制刷新了页面断开了连接)，也得结束
			return false
		}

		// 把大模型吐出来的这个字，立刻推给前端
		c.SSEvent("message", response.Choices[0].Delta.Content)

		return true // 告诉 Gin：我还没发完，继续保持循环！
	})
}
