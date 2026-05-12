package services

// services/leaderboard_service.go

import (
	"context"
	"gojo/global"
	"gojo/models"
	"strconv"
)

// LeaderboardItem 单个玩家的排行信息
type LeaderboardItem struct {
	Rank     int64  `json:"rank"`
	Score    int    `json:"score"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

// LeaderboardData 返回给前端的完整大礼包
type LeaderboardData struct {
	Top50   []LeaderboardItem `json:"top_50"`
	MyRank  int64             `json:"my_rank"`
	MyScore int               `json:"my_score"`
}

// GetGlobalLeaderboard 获取全服排行榜 & 自己的排名
func GetGlobalLeaderboard(ctx context.Context, currentUserID uint) (*LeaderboardData, error) {
	data := &LeaderboardData{
		Top50:  make([]LeaderboardItem, 0),
		MyRank: -1, // 默认没上榜
	}

	// 1. 💥 魔法一：ZREVRANGE WITHSCORES (获取前 50 名卷王)
	topUsersInfo, err := global.Rdb.ZRevRangeWithScores(ctx, "leaderboard:global", 0, 49).Result()
	if err != nil {
		return nil, err
	}

	if len(topUsersInfo) == 0 {
		return data, nil // 榜单还是空的，直接返回
	}

	// 2. 抽离 ID 和 Score
	var topUserIDs []uint
	var topScores []int
	for _, z := range topUsersInfo {
		uid, _ := strconv.Atoi(z.Member.(string))
		topUserIDs = append(topUserIDs, uint(uid))
		topScores = append(topScores, int(z.Score))
	}

	// 3. 极其优雅的 IN 查询与 Map 组装
	var users []models.User
	userMap := make(map[uint]string)
	models.DB.Select("id", "username").Where("id IN ?", topUserIDs).Find(&users)
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	// 4. 恢复 Redis 的绝对顺序，拼装百强榜单
	for i, uid := range topUserIDs {
		// 如果因为某些原因（比如用户被物理删除了），MySQL 里没查到名字，可以给个默认值
		username := userMap[uid]
		if username == "" {
			username = "神秘玩家"
		}

		data.Top50 = append(data.Top50, LeaderboardItem{
			Rank:     int64(i + 1),
			Score:    topScores[i],
			UserID:   uid,
			Username: username,
		})
	}

	// ==========================================
	// 🎁 附加题：查询当前登录用户的“我的排名”
	// ==========================================
	if currentUserID != 0 {
		uidStr := strconv.Itoa(int(currentUserID))

		// 💥 魔法二：极速查我的名次
		rank, err := global.Rdb.ZRevRank(ctx, "leaderboard:global", uidStr).Result()
		if err == nil {
			data.MyRank = rank + 1

			// 💥 魔法三：极速查我的分数
			score, _ := global.Rdb.ZScore(ctx, "leaderboard:global", uidStr).Result()
			data.MyScore = int(score)
		}
		// 如果 err == redis.Nil，说明没上榜，MyRank 保持 -1 不变
	}

	return data, nil
}
