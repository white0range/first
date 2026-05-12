// services/testcase_service.go
package services

import (
	"context"
	"errors"
	"gojo/models"
	"gojo/utils"
	"strconv"
)

// AddTestCase 为指定题目添加测试用例
func AddTestCase(ctx context.Context, problemIDStr string, req models.TestCaseRequest) (*models.TestCase, error) {
	// 1. 安全转换 ID
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		return nil, errors.New("invalid_id")
	}

	// 2. 🛡️ 架构师防御：先确认这道题真不真实存在！
	// 防止因为 MySQL 没有强外键约束，导致用例挂在空气上
	var count int64
	models.DB.Model(&models.Problem{}).Where("id = ?", problemID).Count(&count)
	if count == 0 {
		return nil, errors.New("problem_not_found")
	}

	// 3. 组装干净的、准备下硬盘的 DAO 对象
	testCase := models.TestCase{
		ProblemID:      uint(problemID),
		Input:          req.Input,
		ExpectedOutput: req.ExpectedOutput,
	}

	// 4. 落库保存
	if err := models.DB.Create(&testCase).Error; err != nil {
		return nil, err
	}

	// 💡 注意：因为我们在前面的讨论中，决定了向普通用户隐藏 TestCase，
	// 所以这里【不需要】撕毁题目详情的 Redis 缓存。
	// 但如果你的后台管理系统有专门的“题目用例列表”缓存，这里就需要清空它。

	return &testCase, nil
}

// DeleteTestCase 删除指定的测试用例
func DeleteTestCase(ctx context.Context, caseID string) error {
	// 执行物理删除，并拿到执行结果的载体 (result)
	result := models.DB.Delete(&models.TestCase{}, caseID)

	// 1. 如果有底层 SQL 语法错误或连接断开
	if result.Error != nil {
		return result.Error
	}

	// 2. 🛡️ 架构师级拦截：检查到底有没有数据被删掉！
	if result.RowsAffected == 0 {
		return errors.New("case_not_found")
	}

	// 💡 同上一个接口，因为测试用例对普通用户是隐藏的，
	// 我们不需要去撕毁题目的展示缓存，直接返回成功即可。
	return nil
}

// GetTestCases 获取某道题的所有测试样例
func GetTestCases(ctx context.Context, problemIDStr string, page, limit int) (*models.TestCaseListResponse, error) {
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

	res := &models.TestCaseListResponse{
		Page:  page,
		Limit: limit,
	}

	// 2. 🛡️ 提取公共查询器，死死锁定 problem_id！(修复了雷一)
	query := models.DB.Model(&models.TestCase{}).Where("problem_id = ?", problemID)

	// 3. 查真实总数
	if err := query.Count(&res.Total).Error; err != nil {
		return nil, err
	}

	// 4. 挂载分页神器获取当前页数据
	if err := query.Scopes(utils.Paginate(page, limit)).Find(&res.Items).Error; err != nil {
		return nil, err
	}

	return res, nil
}
