package dto

type JudgeTask struct {
	SubmissionID uint   `json:"submission_id"`
	ProblemID    uint   `json:"problem_id"`
	Code         string `json:"code"`
	UserID       uint   `json:"user_id"`
}
