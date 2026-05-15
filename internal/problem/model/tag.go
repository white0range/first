package model

// Tag 题目标签模型 (比如 "动态规划", "数组")
type Tag struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"` // 标签名必须唯一
}
