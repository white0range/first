// controllers/search_controller.go

package controllers

import (
	"gojo/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SearchProblems 极速复合搜索接口
func SearchProblems(c *gin.Context) {
	// 1. 门卫查参数
	var req services.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	// 2. 呼叫大厅经理 (Service 大脑) 干活
	result, err := services.ExecuteSearch(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索引击异常，请稍后再试"})
		return
	}

	// 3. 完美交卷
	c.JSON(http.StatusOK, gin.H{
		"message": "搜索成功",
		"total":   result.Total,
		"data":    result.Data,
	})
}
