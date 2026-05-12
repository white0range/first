// services/user_service.go
package services

import (
	"context"
	"errors"
	"gojo/models"
	"gojo/utils"
)

// DTO 定义 (建议单独放在 models 或 dto 包，这里为了直观写在一起)
type UserAuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"` // 增加最小长度校验
}

type UserProfileResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Role        int    `json:"role"`
	SolvedCount int    `json:"solved_count"`
	SolvedList  []uint `json:"solved_list"`
}

// RegisterUser 处理用户注册
func RegisterUser(ctx context.Context, req UserAuthRequest) error {
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return errors.New("hash_failed")
	}

	user := models.User{
		Username: req.Username,
		Password: hash,
	}

	// 落库：如果 Username 设了唯一索引，重复会报错
	if err := models.DB.Create(&user).Error; err != nil {
		return errors.New("username_exists")
	}

	return nil
}

// LoginUser 处理登录并签发令牌
func LoginUser(ctx context.Context, req UserAuthRequest) (string, error) {
	var user models.User

	if err := models.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return "", errors.New("user_not_found")
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return "", errors.New("wrong_password")
	}

	token, err := utils.GenerateToken(&user)
	if err != nil {
		return "", errors.New("token_generation_failed")
	}

	return token, nil
}

// GetUserProfile 获取战绩大屏
func GetUserProfile(ctx context.Context, userID uint) (*UserProfileResponse, error) {
	var user models.User
	// 剔除密码等敏感字段
	if err := models.DB.Select("id", "username", "role", "solved_count").First(&user, userID).Error; err != nil {
		return nil, errors.New("user_not_found")
	}

	var solvedProblemIDs []uint
	// 提取 AC 的题目 ID 列表
	models.DB.Model(&models.Submission{}).
		Where("user_id = ? AND status = ?", userID, "AC").
		Distinct("problem_id").
		Pluck("problem_id", &solvedProblemIDs)

	// 组装安全的返回对象
	return &UserProfileResponse{
		ID:          user.ID,
		Username:    user.Username,
		Role:        user.Role,
		SolvedCount: user.SolvedCount,
		SolvedList:  solvedProblemIDs,
	}, nil
}
