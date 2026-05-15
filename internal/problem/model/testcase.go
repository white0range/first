package model

import "gorm.io/gorm"

// TestCase 测试用例表
type TestCase struct {
	gorm.Model

	ProblemID uint `gorm:"index;not null" json:"problem_id"`

	// 判题机需要的弹药
	Input          string `gorm:"type:text;not null" json:"input"`           // 喂给沙箱的输入，比如 "1 2"
	ExpectedOutput string `gorm:"type:text;not null" json:"expected_output"` // 期待的正确输出，比如 "3"
}
