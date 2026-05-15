package repository

import (
	"context"
	"errors"
	"gojo/infrastructure/mysql"
	"gojo/internal/problem/model"
	"gojo/pkg/pagination"
)

type TestCaseRepository interface {
	AddTestCase(ctx context.Context, testCase *model.TestCase) error
	DeleteTestCase(ctx context.Context, caseID string) error
	GetTestCase(ctx context.Context, problemID uint, page, limit int) (int64, []model.TestCase, error)
}

type TestCaseRepoMysql struct{}

func NewTestCaseRepository() TestCaseRepository {
	return &TestCaseRepoMysql{}
}
func (r *TestCaseRepoMysql) AddTestCase(ctx context.Context, testCase *model.TestCase) error {
	// 2. 🛡️ 架构师防御：先确认这道题真不真实存在！
	// 防止因为 MySQL 没有强外键约束，导致用例挂在空气上
	var count int64
	mysql.DB.Model(&model.Problem{}).Where("id = ?", testCase.ProblemID).Count(&count)
	if count == 0 {
		return errors.New("problem_not_found")
	}
	// 4. 落库保存
	return mysql.DB.WithContext(ctx).Create(testCase).Error
}

func (r *TestCaseRepoMysql) DeleteTestCase(ctx context.Context, caseID string) error {
	result := mysql.DB.WithContext(ctx).Where("id = ?", caseID).Delete(&model.TestCase{})

	if result.Error != nil {
		return result.Error // 底层 SQL 报错
	}

	// 🛡️ 检查有没有真实删掉数据，这个动作必须由仓管来做！
	if result.RowsAffected == 0 {
		return errors.New("case_not_found")
	}

	return nil
}

func (r *TestCaseRepoMysql) GetTestCase(ctx context.Context, problemID uint, page, limit int) (int64, []model.TestCase, error) {
	var total int64
	var items []model.TestCase

	// 如果不用事务，直接这样写最清爽：
	query := mysql.DB.WithContext(ctx).Model(&model.TestCase{}).Where("problem_id = ?", problemID)

	// 🚨 注意：不需要判断 query == nil，GORM 的 Model() 永远不会返回 nil

	if err := query.Count(&total).Error; err != nil {
		return 0, nil, err
	}

	if err := query.Scopes(pagination.Paginate(page, limit)).Find(&items).Error; err != nil {
		return 0, nil, err
	}

	return total, items, nil
}
