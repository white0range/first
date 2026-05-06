package controllers

import (
	"gojo/global"
	"gojo/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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
	ctx := c.Request.Context()

	// 1. 💥 Redis 魔法一：ZREVRANGE WITHSCORES (获取前 50 名卷王)
	// ZREVRANGE 表示从大到小排 (Reverse)。0 是第一名，49 是第 50 名。
	topUsersInfo, err := global.Rdb.ZRevRangeWithScores(ctx, "leaderboard:global", 0, 49).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取排行榜失败"})
		return
	}

	// 2. 解析 Redis 返回的数据，顺便去 MySQL 查这些卷王的名字
	var leaderboard []LeaderboardResponse
	var topUserIDs []uint
	var topScore []int

	// 先把 ID 抽出来
	for _, z := range topUsersInfo {
		uid, _ := strconv.Atoi(z.Member.(string))
		topUserIDs = append(topUserIDs, uint(uid))
		topScore = append(topScore, int(z.Score))
	}

	// 去 MySQL 批量查这 50 个人的资料 (极其优雅的 IN 查询，拒绝 for 循环里查数据库！)
	var users []models.User
	userMap := make(map[uint]string)
	if len(topUserIDs) > 0 {
		models.DB.Select("id", "username").Where("id IN ?", topUserIDs).Find(&users)
		for _, u := range users {
			userMap[u.ID] = u.Username
		}
	}

	// 拼装最终的百强榜单发给前端
	for i, z := range topUserIDs {
		leaderboard = append(leaderboard, LeaderboardResponse{
			Rank:     int64(i + 1), // 数组索引从0开始，排名从1开始
			Score:    topScore[i],
			UserID:   z,
			Username: userMap[z],
		})
	}

	// ==========================================
	// 🎁 附加题：查询当前登录用户的“我的排名”
	// ==========================================
	var myRank int64 = -1 // 默认没上榜
	var myScore float64 = 0

	userIDAny, exists := c.Get("userID")
	if exists {
		currentUserID := strconv.Itoa(int(userIDAny.(uint)))

		// 💥 Redis 魔法二：ZREVRANK (极速查我的名次)
		rank, err := global.Rdb.ZRevRank(ctx, "leaderboard:global", currentUserID).Result()
		if err == redis.Nil {
			// 说明他还没做过题，不在榜单里
		} else if err == nil {
			myRank = rank + 1 // Redis 名次从 0 开始算，所以要 +1

			// 💥 Redis 魔法三：ZSCORE (极速查我的分数)
			myScore, _ = global.Rdb.ZScore(ctx, "leaderboard:global", currentUserID).Result()
		}
	}

	// 3. 完美交卷
	c.JSON(http.StatusOK, gin.H{
		"message": "获取榜单成功",
		"data": gin.H{
			"top_50":   leaderboard,
			"my_rank":  myRank,
			"my_score": int(myScore),
		},
	})
}
