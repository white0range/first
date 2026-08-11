package service

import (
	"context"
	"errors"
	"testing"

	"gojo/internal/app/apperror"
	problemModel "gojo/internal/problem/model"
	submissionModel "gojo/internal/submission/model"
)

type failedSubmissionProviderStub struct {
	submission *submissionModel.Submission
	err        error
}

func (s *failedSubmissionProviderStub) GetSubmissionByID(context.Context, string) (*submissionModel.Submission, error) {
	return s.submission, s.err
}

func (s *failedSubmissionProviderStub) GetAllSubmissionsByUserID(context.Context, uint) ([]submissionModel.Submission, error) {
	return nil, nil
}

func (s *failedSubmissionProviderStub) GetRecentFailedSubmissionsByUserID(context.Context, uint, int) ([]submissionModel.Submission, error) {
	return nil, nil
}

type problemCatalogProviderStub struct {
	problem *problemModel.Problem
	err     error
}

func (s *problemCatalogProviderStub) GetAllProblemsWithTags(context.Context) ([]problemModel.Problem, error) {
	return nil, nil
}

func (s *problemCatalogProviderStub) GetProblemByID(context.Context, string) (*problemModel.Problem, error) {
	return s.problem, s.err
}

func TestGetFailedSubmissionDetailReturnsCodeAndProblemContext(t *testing.T) {
	submission := &submissionModel.Submission{
		UserID:       7,
		ProblemID:    40,
		Language:     "go",
		Code:         "package main",
		Status:       "WA",
		ActualOutput: "wrong answer on test case 2",
	}
	submission.ID = 99
	problem := &problemModel.Problem{
		Title:       "Shortest Path",
		Description: "Find the shortest path.",
		Tags:        []problemModel.Tag{{Name: "graph"}},
	}
	problem.ID = 40

	svc := NewChatService(nil, nil, &failedSubmissionProviderStub{submission: submission}, &problemCatalogProviderStub{problem: problem})
	detail, err := svc.GetFailedSubmissionDetail(context.Background(), 7, 99)
	if err != nil {
		t.Fatalf("GetFailedSubmissionDetail returned error: %v", err)
	}
	if detail.Code != submission.Code || detail.ProblemDescription != problem.Description {
		t.Fatalf("unexpected detail: %+v", detail)
	}
	if len(detail.ProblemTags) != 1 || detail.ProblemTags[0] != "graph" {
		t.Fatalf("unexpected problem tags: %v", detail.ProblemTags)
	}
}

func TestGetFailedSubmissionDetailRejectsAnotherUser(t *testing.T) {
	submission := &submissionModel.Submission{UserID: 8, ProblemID: 40, Status: "WA"}
	submission.ID = 99
	svc := NewChatService(nil, nil, &failedSubmissionProviderStub{submission: submission}, &problemCatalogProviderStub{})

	_, err := svc.GetFailedSubmissionDetail(context.Background(), 7, 99)
	if !errors.Is(err, apperror.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}
