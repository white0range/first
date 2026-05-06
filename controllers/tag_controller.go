package controllers

import (
	"gojo/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetTagList 获取全站标签列表 (公共接口，无需手环)
// 前端拿这个接口的数据去画页面的“标签筛选栏”
func GetTagList(c *gin.Context) {
	var tags []models.Tag
	// 全量查出所有标签。因为标签一般也就几十上百个，不需要分页，直接 Find 即可
	models.DB.Find(&tags)

	c.JSON(http.StatusOK, gin.H{
		"message": "获取标签列表成功",
		"data":    tags,
	})
}

// CreateTag 超管创建新标签 (受禁卫军保护)
func CreateTag(c *gin.Context) {
	// 定义一个极简的接收结构体，只需要名字
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标签名称不能为空！"})
		return
	}

	tag := models.Tag{Name: req.Name}

	// 插入数据库
	// 注意：因为你在模型里加了 uniqueIndex，如果超管建了两个一模一样的标签，GORM 会自动报错拦截！
	if err := models.DB.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败，该标签可能已存在！"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "标签创建成功",
		"data":    tag,
	})
}

// DeleteTag 删除标签（会同步清除题目关联）
func DeleteTag(c *gin.Context) {
	tagID := c.Param("id")

	var tag models.Tag
	if err := models.DB.First(&tag, tagID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "该标签不存在！"})
		return
	}

	// 💥 GORM 魔法一：级联删除
	// 加上 Select("Problems") 后，GORM 不仅会删除 tags 表里的记录，
	// 还会极其聪明地去 problem_tags 中间表里，把所有包含该 tag_id 的行全部删掉！
	if err := models.DB.Select("Problems").Delete(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除标签失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "标签及其关联已彻底删除"})
}
