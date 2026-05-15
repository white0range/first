package service

import (
	"context"
	"gojo/internal/leaderboard/dto"
	"gojo/internal/leaderboard/repository"
)

// 🚨 大厂魔法：排行榜的“外交官”
// 排行榜 Service 表示：我不管谁去查数据库，给我一个能把 []ID 变成 Map 的工具就行！
type UserProvider interface {
	GetUsersMapByIDs(ctx context.Context, userIDs []uint) (map[uint]string, error)
}

// 1. 结构体注入
type LeaderboardService struct {
	repo         repository.LeaderboardRepository
	userProvider UserProvider // 👈 注入隔壁 User 模块的实力
}

func NewLeaderboardService(r repository.LeaderboardRepository, up UserProvider) *LeaderboardService {
	return &LeaderboardService{repo: r, userProvider: up}
}

// GetGlobalLeaderboard 获取全服排行榜 & 自己的排名
func (s *LeaderboardService) GetGlobalLeaderboard(ctx context.Context, currentUserID uint) (*dto.LeaderboardData, error) {
	data := &dto.LeaderboardData{
		Top50:  make([]dto.LeaderboardItem, 0),
		MyRank: -1,
	}

	// 1. 呼叫仓管，拿到极其干净的 Top 50 数据记录
	records, err := s.repo.GetTopN(ctx, 50)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return data, nil
	}

	// 2. 提取出所有的 UserID
	var topUserIDs []uint
	for _, rec := range records {
		topUserIDs = append(topUserIDs, rec.UserID)
	}

	// 3. 呼叫外交官去隔壁拿名字！(屏蔽了 MySQL 查询)
	userMap := make(map[uint]string)
	if s.userProvider != nil {
		userMap, _ = s.userProvider.GetUsersMapByIDs(ctx, topUserIDs)
	}

	// 4. 拼装百强榜单发给前端
	for i, rec := range records {
		username := userMap[rec.UserID]
		if username == "" {
			username = "神秘玩家"
		}

		data.Top50 = append(data.Top50, dto.LeaderboardItem{
			Rank:     int64(i + 1),
			Score:    rec.Score,
			UserID:   rec.UserID,
			Username: username,
		})
	}

	// 5. 附加题：查询当前登录用户的“我的排名”
	if currentUserID != 0 {
		rank, score, err := s.repo.GetUserRankAndScore(ctx, currentUserID)
		if err == nil && rank != -1 {
			data.MyRank = rank
			data.MyScore = score
		}
	}

	return data, nil
}
