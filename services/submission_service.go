// services/submission_service.go
package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gojo/global"
	"gojo/models"
	"gojo/utils"

	"github.com/sashabaranov/go-openai"
)

// SubmitCode 将提交记录落库，并推送到 Redis 判题队列
func SubmitCode(ctx context.Context, userID uint, req models.SubmitRequest) (*models.Submission, error) {
	// 1. 组装实体，准备落库
	submission := models.Submission{
		UserID:    userID,
		ProblemID: req.ProblemID,
		Language:  req.Language,
		Code:      req.Code,
		// Status 不需要写，GORM 自动设为默认值 'Pending'
	}

	// 2. 落库保存！
	if err := models.DB.Create(&submission).Error; err != nil {
		return nil, fmt.Errorf("入库失败: %w", err)
	}

	// 3. 打包给判题工人的任务参数
	task := map[string]interface{}{
		"user_id":       userID,
		"submission_id": submission.ID,
		"problem_id":    req.ProblemID,
		"code":          req.Code,
	}

	taskBytes, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("任务序列化失败: %w", err)
	}

	// 4. 狠狠地塞进 Redis 传送带！
	// 💡 细节：这里直接复用了传进来的 ctx，保持链路完整
	if err := global.Rdb.LPush(ctx, "judge_queue", taskBytes).Err(); err != nil {
		return nil, fmt.Errorf("推送判题队列失败: %w", err)
	}

	return &submission, nil
}

// GetSubmissionResult 获取提交记录（带水平越权校验）
func GetSubmissionResult(ctx context.Context, submissionID string, currentUserID uint) (*models.Submission, error) {
	var submission models.Submission

	// 1. 去数据库里查这条记录
	if err := models.DB.First(&submission, submissionID).Error; err != nil {
		return nil, err // 交给 Controller 处理 404
	}

	// 2. 🛡️ 架构师防御：水平越权校验 (Horizontal Privilege Escalation)
	// 绝对不能让别人看到你的代码！除非你是管理员（如果你有管理员字段也可以加在这里）
	if submission.UserID != currentUserID {
		return nil, errors.New("unauthorized access") // 发现偷窥，直接拦截！
	}

	return &submission, nil
}

// services/submission_service.go

// MySubmissionsResponse 专属的数据传输对象 (DTO)
type MySubmissionsResponse struct {
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
	Items []models.Submission `json:"items"`
}

// GetMySubmissions 获取指定用户的提交历史
func GetMySubmissions(ctx context.Context, userID uint, page, limit int) (*MySubmissionsResponse, error) {
	// 参数兜底
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	res := &MySubmissionsResponse{
		Page:  page,
		Limit: limit,
	}

	// 1. 提取出一个公共的 query 对象，死死锁定当前用户的 ID
	query := models.DB.Model(&models.Submission{}).Where("user_id = ?", userID)

	// 2. 查总数 (没有任何 Limit 和 Offset 的纯净查询)
	if err := query.Count(&res.Total).Error; err != nil {
		return nil, err
	}

	// 3. 挂载解耦后的分页神器 & 排序剔除 & 获取数据
	// 💥 注意看这里：使用的是我们重构后只接收 page, limit 的纯粹分页工具
	err := query.Scopes(utils.Paginate(page, limit)).
		Order("created_at desc").      // 最新提交排前面
		Omit("code", "actual_output"). // 剔除大体积字段，保护带宽
		Find(&res.Items).Error

	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetAIAssistanceStream 负责核心业务校验，并向大模型发起连接，返回数据流
func GetAIAssistanceStream(ctx context.Context, submissionID string, userID uint) (*openai.ChatCompletionStream, error) {
	var submission models.Submission

	// 1. 查出这条战绩
	if err := models.DB.First(&submission, submissionID).Error; err != nil {
		return nil, errors.New("not_found")
	}

	// 2. 🛡️ 极其关键的安全边界：水平越权校验
	if submission.UserID != userID {
		return nil, errors.New("forbidden")
	}

	// 3. 🚦 业务规则：AC 的代码不予诊断
	if submission.Status == "AC" {
		return nil, errors.New("already_ac")
	}

	// 4. 召唤 AI，拿到水龙头
	// 💡 架构师建议：如果是大厂标准，AskAIStream 最好也接收一个 ctx，这样用户如果关掉浏览器，大模型那边也能瞬间停止生成，省钱！
	stream, err := utils.AskAIStream(submission.Code, submission.Language, submission.ActualOutput)
	if err != nil {
		return nil, errors.New("ai_connect_failed")
	}

	// 5. 把没关紧的水龙头直接扔给 Controller
	return stream, nil
}
