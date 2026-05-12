package models

import "gorm.io/gorm"

// TestCase 测试用例表
type TestCase struct {
	gorm.Model

	ProblemID uint `gorm:"index;not null" json:"problem_id"`

	// 判题机需要的弹药
	Input          string `gorm:"type:text;not null" json:"input"`           // 喂给沙箱的输入，比如 "1 2"
	ExpectedOutput string `gorm:"type:text;not null" json:"expected_output"` // 期待的正确输出，比如 "3"
}

// =====================================
// TestCaseRequest 测试用例接客表单 (DTO)
// 专门用于接收前端传来的单个测试用例数据
// =====================================
type TestCaseRequest struct {
	Input          string `json:"input" binding:"required"`
	ExpectedOutput string `json:"expected_output" binding:"required"`
}

// TestCaseListResponse 专属的列表返回对象
type TestCaseListResponse struct {
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
	Items []TestCase `json:"items"`
}
