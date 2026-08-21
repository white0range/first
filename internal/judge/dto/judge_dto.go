package dto

import (
	"fmt"
	"time"
)

type JudgeTask struct {
	ID           string    `json:"id"`
	SubmissionID uint      `json:"submission_id"`
	ProblemID    uint      `json:"problem_id"`
	Code         string    `json:"code"`
	UserID       uint      `json:"user_id"`
	RetryCount   int       `json:"retry_count"`
	CreatedAt    time.Time `json:"created_at"`
	LastError    string    `json:"last_error,omitempty"`
}

func NewJudgeTask(submissionID, problemID, userID uint, code string) JudgeTask {
	now := time.Now().UTC()
	return JudgeTask{
		ID:           fmt.Sprintf("judge-%d-%d", submissionID, now.UnixNano()),
		SubmissionID: submissionID,
		ProblemID:    problemID,
		Code:         code,
		UserID:       userID,
		CreatedAt:    now,
	}
}

func (t JudgeTask) Valid() bool {
	return t.ID != "" && t.SubmissionID > 0 && t.ProblemID > 0 && t.UserID > 0 && t.Code != ""
}
