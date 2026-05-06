package utils

import (
	"context"
	"gojo/global"
	"strconv"
)

const GlobalLeaderboardKey = "leaderboard:global"

// AddUserScore 给玩家加分（当他 AC 题目后调用）
func AddUserScore(userID uint, scoreToAdd float64) error {
	ctx := context.Background()
	// 把 userID 转成字符串，作为 ZSet 的 Member
	member := strconv.Itoa(int(userID))

	// 💥 核心魔法：ZINCRBY
	// 如果玩家不存在，Redis 会自动把他加进去并设为该分数；如果存在，就在原分数上累加！
	_, err := global.Rdb.ZIncrBy(ctx, GlobalLeaderboardKey, scoreToAdd, member).Result()
	return err
}
