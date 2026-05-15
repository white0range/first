package handler

import (
	"gojo/internal/problem/dto"
	"gojo/internal/problem/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TestCaseHandler struct {
	svc *service.TestCaseService
}

func NewTestCaseHandler(s *service.TestCaseService) *TestCaseHandler {
	return &TestCaseHandler{svc: s}
}

// AddTestCase 为某道题目单独添加一个测试样例
func (h *TestCaseHandler) AddTestCase(c *gin.Context) {
	problemIDStr := c.Param("id")

	// 1. 使用专属的 dto 接收参数，彻底杜绝 Over-Posting
	var req dto.TestCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误，必须包含 input 和 expected_output"})
		return
	}

	// 2. 呼叫 Service 大脑
	id, err := h.svc.AddTestCase(c.Request.Context(), problemIDStr, req)
	if err != nil {
		// 精准拦截不同类型的错误
		switch err.Error() {
		case "invalid_id":
			c.JSON(http.StatusBadRequest, gin.H{"error": "题目 ID 格式错误"})
		case "problem_not_found":
			c.JSON(http.StatusNotFound, gin.H{"error": "操作失败：该题目不存在！"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，测试用例添加失败"})
		}
		return
	}

	// 3. 返回喜报
	c.JSON(http.StatusOK, gin.H{
		"message": "测试用例添加成功",
		"case_id": id,
	})
}

// DeleteTestCase 删除某个指定的测试样例
func (h *TestCaseHandler) DeleteTestCase(c *gin.Context) {
	// 1. 获取 URL 里的样例 ID
	caseID := c.Param("case_id")

	// 2. 呼叫 Service 执行抹杀命令
	if err := h.svc.DeleteTestCase(c.Request.Context(), caseID); err != nil {
		if err.Error() == "case_not_found" {
			// 完美化解“空删欺骗”，精准报错
			c.JSON(http.StatusNotFound, gin.H{"error": "斩杀失败：该测试用例不存在或已被删除！"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，测试用例删除失败"})
		return
	}

	// 3. 返回成功
	c.JSON(http.StatusOK, gin.H{
		"message": "测试用例已成功物理删除！",
	})
}

// GetTestCases 获取某道题的所有测试样例（管理员专属，严禁外泄！）
func (h *TestCaseHandler) GetTestCases(c *gin.Context) {
	// 1. 取出路径上的题目 ID
	problemIDStr := c.Param("id")

	// 2. 提取分页参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	// 3. 呼叫 Service 大脑
	res, err := h.svc.GetTestCases(c.Request.Context(), problemIDStr, page, limit)
	if err != nil {
		if err.Error() == "invalid_id" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "题目 ID 格式错误"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，获取样例列表失败"})
		return
	}

	// 4. 返回标准格式 (修复了雷二，res 里面已经带有真实的 Total)
	c.JSON(http.StatusOK, gin.H{
		"message": "获取样例列表成功",
		"data":    res,
	})
}
