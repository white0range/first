// services/testcase_service.go
package service

import (
	"context"
	"errors"
	"gojo/internal/problem/dto"
	"gojo/internal/problem/model"
	"gojo/internal/problem/repository"
	"strconv"
)

type TestCaseService struct {
	repo repository.TestCaseRepository
}

func NewTestCaseService(r repository.TestCaseRepository) *TestCaseService {
	return &TestCaseService{repo: r}
}

// AddTestCase 为指定题目添加测试用例
func (s *TestCaseService) AddTestCase(ctx context.Context, problemIDStr string, req dto.TestCaseRequest) (uint, error) {
	// 1. 安全转换 ID
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		return 0, errors.New("invalid_id")
	}

	// 3. 组装干净的、准备下硬盘的 DAO 对象
	testCase := model.TestCase{
		ProblemID:      uint(problemID),
		Input:          req.Input,
		ExpectedOutput: req.ExpectedOutput,
	}

	if err := s.repo.AddTestCase(ctx, &testCase); err != nil {
		return 0, err
	}

	// 💡 注意：因为我们在前面的讨论中，决定了向普通用户隐藏 TestCase，
	// 所以这里【不需要】撕毁题目详情的 Redis 缓存。
	// 但如果你的后台管理系统有专门的“题目用例列表”缓存，这里就需要清空它。

	return testCase.ID, nil
}

// DeleteTestCase 删除指定的测试用例
func (s *TestCaseService) DeleteTestCase(ctx context.Context, caseID string) error {
	// Service 只需要负责发号施令，不用去拿什么 result 载体了
	return s.repo.DeleteTestCase(ctx, caseID)
}

// GetTestCases 获取某道题的所有测试样例
func (s *TestCaseService) GetTestCases(ctx context.Context, problemIDStr string, page, limit int) (*dto.TestCaseListResponse, error) {
	// 1. 安全转换 ID
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		return nil, errors.New("invalid_id")
	}

	// 参数兜底
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	res := &dto.TestCaseListResponse{
		Page:  page,
		Limit: limit,
	}
	res.Total, res.Items, err = s.repo.GetTestCase(ctx, uint(problemID), page, limit)
	return res, err
}
