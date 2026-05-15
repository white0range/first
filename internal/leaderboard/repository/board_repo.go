package repository

import (
	"context"
	"errors"
	"gojo/infrastructure/cache"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// 1. 定义极其干净的内部载体，绝不向外泄露 redis.Z
type RankRecord struct {
	UserID uint
	Score  int
}

// 2. 仓管接口定义
type LeaderboardRepository interface {
	GetTopN(ctx context.Context, limit int64) ([]RankRecord, error)
	GetUserRankAndScore(ctx context.Context, userID uint) (rank int64, score int, err error)
}

type leaderboardRepoRedis struct {
	// 这里的 key 可以做成配置化的，现在先写死
	leaderboardKey string
}

func NewLeaderboardRepository() LeaderboardRepository {
	return &leaderboardRepoRedis{leaderboardKey: "leaderboard:infrastructure"}
}

// 3. 落地实现
func (r *leaderboardRepoRedis) GetTopN(ctx context.Context, limit int64) ([]RankRecord, error) {
	// 底层苦逼地查 Redis
	zs, err := cache.Rdb.ZRevRangeWithScores(ctx, r.leaderboardKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	// 🛡️ 架构师防线：在仓管这里就把 redis.Z 洗干净，变成纯粹的 RankRecord 交上去！
	var records []RankRecord
	for _, z := range zs {
		uid, _ := strconv.Atoi(z.Member.(string))
		records = append(records, RankRecord{
			UserID: uint(uid),
			Score:  int(z.Score),
		})
	}
	return records, nil
}

func (r *leaderboardRepoRedis) GetUserRankAndScore(ctx context.Context, userID uint) (int64, int, error) {
	uidStr := strconv.Itoa(int(userID))

	rank, err := cache.Rdb.ZRevRank(ctx, r.leaderboardKey, uidStr).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return -1, 0, nil // 没上榜不是报错，是正常业务逻辑
		}
		return -1, 0, err
	}

	score, _ := cache.Rdb.ZScore(ctx, r.leaderboardKey, uidStr).Result()
	return rank + 1, int(score), nil // 返回真实的名次（+1）
}
