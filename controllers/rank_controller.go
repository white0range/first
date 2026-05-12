package controllers

import (
	"gojo/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LeaderboardResponse 专门给前端返回的排行榜结构体
type LeaderboardResponse struct {
	Rank  int64 `json:"rank"`  // 排名 (第1名、第2名...)
	Score int   `json:"score"` // 战力值
	// 下面这俩需要去 MySQL 里查
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

// GetGlobalLeaderboard 获取全服排行榜 & 自己的排名
func GetGlobalLeaderboard(c *gin.Context) {
	// 1. 🛡️ 安全探查身份 (因为这是公共接口，游客也能看排行榜)
	var currentUserID uint = 0
	if userIDRaw, exists := c.Get("userID"); exists {
		if uid, ok := userIDRaw.(uint); ok {
			currentUserID = uid
		}
	}

	// 2. 呼叫 Service 大脑
	data, err := services.GetGlobalLeaderboard(c.Request.Context(), currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，获取排行榜失败"})
		return
	}

	// 3. 完美交卷
	c.JSON(http.StatusOK, gin.H{
		"message": "获取榜单成功",
		"data":    data,
	})
}
