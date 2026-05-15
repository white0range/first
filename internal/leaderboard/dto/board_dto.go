package dto

// LeaderboardResponse 专门给前端返回的排行榜结构体
type LeaderboardResponse struct {
	Rank  int64 `json:"leaderboard"` // 排名 (第1名、第2名...)
	Score int   `json:"score"`       // 战力值
	// 下面这俩需要去 MySQL 里查
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
}

// LeaderboardItem 单个玩家的排行信息
type LeaderboardItem struct {
	Rank     int64  `json:"leaderboard"`
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
