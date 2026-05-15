package dto

import "gojo/internal/problem/model"

// =====================================
// TestCaseRequest 测试用例接客表单 (dto)
// 专门用于接收前端传来的单个测试用例数据
// =====================================
type TestCaseRequest struct {
	Input          string `json:"input" binding:"required"`
	ExpectedOutput string `json:"expected_output" binding:"required"`
}

// TestCaseListResponse 专属的列表返回对象
type TestCaseListResponse struct {
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
	Items []model.TestCase `json:"items"`
}
