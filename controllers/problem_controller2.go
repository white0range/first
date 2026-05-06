package controllers

import (
	"gojo/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UpdateProblemTagsRequest 接收前端传来的全新标签 ID 数组
type UpdateProblemTagsRequest struct {
	TagIDs []uint `json:"tag_ids" binding:"required"`
}

// UpdateProblemTags 更新题目的标签（全量替换）
func UpdateProblemTags(c *gin.Context) {
	problemID := c.Param("id")
	req := UpdateProblemTagsRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	// 1. 找到这道题
	var problem models.Problem
	if err := models.DB.First(&problem, problemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
		return
	}

	// 2. 把前端传来的新标签 ID，去数据库换成真实的 Tag 实体
	var tags []models.Tag
	if len(req.TagIDs) > 0 {
		models.DB.Find(&tags, req.TagIDs)
	}

	// 💥 GORM 魔法二：关联替换 (Replace)
	// 这行代码极其狂暴：它会自动分析 problem 现有的标签，和传进来的 tags 进行比对。
	// 把不要的从中间表删除，把新增的添加到中间表。全程只需要这一行代码！
	if err := models.DB.Model(&problem).Association("Tags").Replace(tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新标签关联失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "题目标签更新成功"})
}
