package service

import (
	"context"
	"errors"
	"gojo/internal/user/dto" // 建议把 UserAuthRequest 放这里
	"gojo/internal/user/model"
	"gojo/internal/user/repository"
	"gojo/pkg/jwt"
	"gojo/pkg/password"
)

// 🚨：定义一个“外交接口”
// 用户 Service ：我不管你是谁，只要你能给我提供用户的 AC 列表就行！
type SubmissionProvider interface {
	GetACProblemIDsByUserID(ctx context.Context, userID uint) ([]uint, error)
}

// 1. 组装 Service 结构体
type UserService struct {
	repo        repository.UserRepository
	subProvider SubmissionProvider // 👈 跨域依赖注入
}

// 2. 构造函数
func NewUserService(r repository.UserRepository, sp SubmissionProvider) *UserService {
	return &UserService{
		repo:        r,
		subProvider: sp,
	}
}

// RegisterUser 处理用户注册
func (s *UserService) RegisterUser(ctx context.Context, req dto.UserAuthRequest) error {
	hash, err := password.HashPassword(req.Password)
	if err != nil {
		return errors.New("hash_failed")
	}

	user := model.User{
		Username: req.Username,
		Password: hash,
	}

	// 呼叫仓管
	if err := s.repo.CreateUser(ctx, &user); err != nil {
		return errors.New("username_exists")
	}

	return nil
}

// LoginUser 处理登录并签发令牌
func (s *UserService) LoginUser(ctx context.Context, req dto.UserAuthRequest) (string, error) {
	// 呼叫仓管拿数据
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return "", errors.New("user_not_found")
	}

	// 纯业务逻辑
	if !password.CheckPasswordHash(req.Password, user.Password) {
		return "", errors.New("wrong_password")
	}

	token, err := jwt.GenerateToken(user)
	if err != nil {
		return "", errors.New("token_generation_failed")
	}

	return token, nil
}

// GetUserProfile 获取战绩大屏
func (s *UserService) GetUserProfile(ctx context.Context, userID uint) (*dto.UserProfileResponse, error) {
	// 1. 呼叫自己家的仓管
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user_not_found")
	}

	// 2. 呼叫“外交官”拿隔壁的 AC 数据！彻底告别直接查 submission 表！
	var solvedProblemIDs []uint
	if s.subProvider != nil {
		solvedProblemIDs, _ = s.subProvider.GetACProblemIDsByUserID(ctx, userID)
	}

	// 3. 组装安全的返回对象
	return &dto.UserProfileResponse{
		ID:          user.ID,
		Username:    user.Username,
		Role:        user.Role,
		SolvedCount: user.SolvedCount,
		SolvedList:  solvedProblemIDs,
	}, nil
}

// GetUsersMapByIDs 提供给其他模块（如排行榜）的跨域调用接口
// 将查询到的用户列表转化为 O(1) 查询亮度的 Map
func (s *UserService) GetUsersMapByIDs(ctx context.Context, userIDs []uint) (map[uint]string, error) {
	// 1. 初始化一个空 Map 兜底
	userMap := make(map[uint]string)

	// 2. 如果排行榜还没人（空数组），直接返回空 Map
	if len(userIDs) == 0 {
		return userMap, nil
	}

	// 3. 呼叫自己家的小弟（仓管）去查数据库
	users, err := s.repo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	// 4. 组装成 Map
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	return userMap, nil
}
