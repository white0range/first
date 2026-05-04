package controllers

import (
	"gojo/models"
	"gojo/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AddTestCase 为某道题目单独添加一个测试样例
func AddTestCase(c *gin.Context) {
	// 获取 URL 里的题目 ID (/problems/:id/cases)
	problemIDStr := c.Param("id")
	problemID, _ := strconv.Atoi(problemIDStr)

	// 接收前端传来的单个样例数据 ({"input": "1 2", "expected_output": "3"})
	var req models.TestCase
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误，必须包含 input 和 expected_output"})
		return
	}

	// 强制打上所属题目的烙印
	req.ProblemID = uint(problemID)

	// 存入数据库
	if err := models.DB.Create(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，测试用例添加失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "测试用例添加成功",
		"case_id": req.ID, // 返回新生成的样例 ID，方便以后删除
	})
}

// DeleteTestCase 删除某个指定的测试样例
func DeleteTestCase(c *gin.Context) {
	// 获取 URL 里的样例 ID (/problems/cases/:case_id)
	caseID := c.Param("case_id")

	// 直接从 test_cases 表里精准抹杀！
	// 相当于 DELETE FROM test_cases WHERE id = ?
	if err := models.DB.Delete(&models.TestCase{}, caseID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，测试用例删除失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "测试用例已成功物理删除！"})
}

// GetTestCases 获取某道题的所有测试样例（管理员专属，严禁外泄！）
func GetTestCases(c *gin.Context) {
	// 获取 URL 里的题目 ID (/problems/:id/cases)
	problemID := c.Param("id")

	var testCases []models.TestCase
	var total int64
	// 让 GORM 去底层捞数据：SELECT * FROM test_cases WHERE problem_id = ?
	models.DB.Model(&models.TestCase{}).Count(&total)
	if err := models.DB.Where("problem_id = ?", problemID).Scopes(utils.Paginate(c)).Find(&testCases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，获取样例列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取样例列表成功",
		"total":   len(testCases), // 顺便告诉前端一共有几个
		"data":    testCases,
	})
}
