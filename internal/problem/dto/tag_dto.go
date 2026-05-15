package dto

// CreateTagRequest 定义 dto (建议放 models 里)
type CreateTagRequest struct {
	Name string `json:"name" binding:"required"`
}
