package handler

import (
	"gojo/internal/problem/dto"
	"gojo/internal/problem/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TagHandler struct {
	svc *service.TagService
}

func NewTagHandler(svc *service.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

// GetTagList 获取全站标签列表 (公共接口，无需手环)
func (h *TagHandler) GetTagList(c *gin.Context) {
	tags, err := h.svc.GetTagList(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取标签列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取标签列表成功",
		"data":    tags,
	})
}

// CreateTag 超管创建新标签 (受禁卫军保护)
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req dto.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "标签名称不能为空！"})
		return
	}

	tag, err := h.svc.CreateTag(c.Request.Context(), req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败，该标签可能已存在！"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "标签创建成功",
		"data":    tag,
	})
}

// DeleteTag 删除标签（会同步清除题目关联并撕毁缓存）
func (h *TagHandler) DeleteTag(c *gin.Context) {
	tagID := c.Param("id")

	if err := h.svc.DeleteTag(c.Request.Context(), tagID); err != nil {
		// 精准拦截 NotFound 错误
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "该标签不存在！"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除标签失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "标签及其关联已彻底删除"})
}
