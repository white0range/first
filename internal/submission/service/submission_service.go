package service

import (
	"context"
	"fmt"

	"gojo/internal/app/apperror"
	judgeDTO "gojo/internal/judge/dto"
	judgeQueue "gojo/internal/judge/queue"
	"gojo/internal/submission/dto"
	"gojo/internal/submission/model"
	"gojo/internal/submission/repository"
)

type SubmissionService struct {
	repo  repository.SubmissionRepository
	queue *judgeQueue.Queue
}

func NewSubmissionService(r repository.SubmissionRepository) *SubmissionService {
	return &SubmissionService{repo: r, queue: judgeQueue.New()}
}

func (s *SubmissionService) SubmitCode(ctx context.Context, userID uint, req dto.SubmitRequest) (*model.Submission, error) {
	submission := model.Submission{
		UserID:    userID,
		ProblemID: req.ProblemID,
		Language:  req.Language,
		Code:      req.Code,
	}

	if err := s.repo.CreateSubmission(ctx, &submission); err != nil {
		return nil, fmt.Errorf("create submission failed: %w", err)
	}

	task := judgeDTO.NewJudgeTask(submission.ID, req.ProblemID, userID, req.Code)
	if _, err := s.queue.Enqueue(ctx, task); err != nil {
		// Keep the record Pending so the worker's MySQL reconciler can enqueue it
		// after a temporary Redis outage instead of losing the submission.
		return nil, fmt.Errorf("enqueue judge task failed: %w", err)
	}

	return &submission, nil
}

func (s *SubmissionService) GetSubmissionResult(ctx context.Context, submissionID string, currentUserID uint) (*model.Submission, error) {
	submission, err := s.repo.GetSubmissionByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}

	if submission.UserID != currentUserID {
		return nil, apperror.ErrUnauthorizedAccess
	}

	return submission, nil
}

func (s *SubmissionService) GetMySubmissions(ctx context.Context, userID uint, page, limit int) (*dto.MySubmissionsResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	total, items, err := s.repo.GetSubmissionsByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	return &dto.MySubmissionsResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Items: items,
	}, nil
}

func (s *SubmissionService) GetACProblemIDsByUserID(ctx context.Context, userID uint) ([]uint, error) {
	return s.repo.GetACProblemIDsByUserID(ctx, userID)
}
