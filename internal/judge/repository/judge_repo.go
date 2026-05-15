package repository

import (
	"context"
	"gojo/infrastructure/mysql"
	"gojo/internal/problem/model"
	submodel "gojo/internal/submission/model"
	usermodel "gojo/internal/user/model"
	"gojo/pkg/addscore"

	"gorm.io/gorm"
)

type JudgeRepository interface {
	GetProblemWithCases(ctx context.Context, problemID uint) (*model.Problem, error)
	UpdateJudgeResult(ctx context.Context, subID, problemID, userID uint, status, output string) error
}

type judgeRepoMysql struct{}

func NewJudgeRepository() JudgeRepository {
	return &judgeRepoMysql{}
}

func (r *judgeRepoMysql) GetProblemWithCases(ctx context.Context, problemID uint) (*model.Problem, error) {
	var problem model.Problem
	err := mysql.DB.WithContext(ctx).Preload("TestCases").First(&problem, problemID).Error
	return &problem, err
}

// UpdateJudgeResult 终极打包：更新提交记录、统计数据、计算积分 (自带事务保护！)
func (r *judgeRepoMysql) UpdateJudgeResult(ctx context.Context, subID, problemID, userID uint, status, output string) error {
	return mysql.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 更新本次提交的结果
		if err := tx.Model(&submodel.Submission{}).Where("id = ?", subID).Updates(map[string]interface{}{
			"status":        status,
			"actual_output": output,
		}).Error; err != nil {
			return err
		}

		// 2. 题目总提交数 +1
		tx.Model(&model.Problem{}).Where("id = ?", problemID).UpdateColumn("submit_count", gorm.Expr("submit_count + ?", 1))

		// 3. 如果 AC，进行后续结算
		if status == "AC" {
			// 题目通过数 +1
			tx.Model(&model.Problem{}).Where("id = ?", problemID).UpdateColumn("accepted_count", gorm.Expr("accepted_count + ?", 1))

			// 查历史 AC 记录防刷分
			var acCount int64
			tx.Model(&submodel.Submission{}).Where("user_id = ? AND problem_id = ? AND status = 'AC'", userID, problemID).Count(&acCount)

			// 如果是首次 AC (当前这条记录已经存进去了，所以是 == 1)
			if acCount == 1 {
				tx.Model(&usermodel.User{}).Where("id = ?", userID).UpdateColumn("solved_count", gorm.Expr("solved_count + ?", 1))
				// 给 Redis 加分
				_ = addscore.AddUserScore(userID, 10.0)
			}
		}
		return nil
	})
}
