package dto

import (
	"gojo/internal/problem/model"
)

// 顺便在这里准备一个接客用的 dto 表单
// 因为我们不希望管理员在发布题目时，还要手填 SubmitCount 这种应该由系统管理的字段
type ProblemRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	TimeLimit   int    `json:"time_limit"` // 如果前端不传，在 Controller 里我们会给它赋默认值
	MemoryLimit int    `json:"memory_limit"`

	// 直接复用 TestCaseRequest！
	TestCases []TestCaseRequest `json:"test_cases"`
	TagIDs    []uint            `json:"tag_ids"`
}

type UpdateProblemTagsRequest struct {
	TagIDs []uint `json:"tag_ids"`
}

type ProblemListResponse struct {
	Total   int64           `json:"total"`
	Page    int             `json:"page"`
	Limit   int             `json:"limit"`
	TagID   string          `json:"tag_id"`
	Message string          `json:"message"`
	Items   []model.Problem `json:"items"`
}
