package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gojo/infrastructure/cache"
	"gojo/internal/problem/model"
	"gojo/internal/problem/repository"
	"time"
)

type TagService struct {
	repo repository.TagRepository
}

func NewTagService(r repository.TagRepository) *TagService {
	return &TagService{repo: r}
}

const TagCacheKey = "cache:tags:all" // 全站标签只有一份，固定 Key

// GetTagList 获取全站标签（走缓存）
func (s *TagService) GetTagList(ctx context.Context) ([]model.Tag, error) {
	var tags []model.Tag

	// 1. 先问 Redis 要
	cachedData, err := cache.Rdb.Get(ctx, TagCacheKey).Result()
	if err == nil {
		fmt.Println("⚡ 触发 Redis 缓存！全站标签极速返回！")
		err := json.Unmarshal([]byte(cachedData), &tags)
		if err != nil {
			return nil, err
		}
		return tags, nil
	}

	err = s.repo.GetTagList(ctx, &tags)

	// 3. 查完塞给 Redis (标签很久才变一次，直接给个 7 天过期时间)
	jsonBytes, _ := json.Marshal(tags)
	cache.Rdb.Set(ctx, TagCacheKey, jsonBytes, 7*24*time.Hour)

	return tags, nil
}

// CreateTag 创建标签并撕毁标签缓存
func (s *TagService) CreateTag(ctx context.Context, name string) (*model.Tag, error) {
	tag := model.Tag{Name: name}

	// 插入数据库 (如果 name 有 Unique 索引，重复会报错)
	if err := s.repo.CreateTag(ctx, &tag); err != nil {
		return nil, err
	}

	// 🚨 撕毁标签缓存，下次获取就会拿最新版
	cache.Rdb.Del(ctx, TagCacheKey)

	return &tag, nil
}

// DeleteTag 删除标签，级联清理中间表，并触发“核弹级”缓存撕毁
func (s *TagService) DeleteTag(ctx context.Context, tagID string) error {

	err := s.repo.DeleteTag(ctx, tagID)
	if err != nil {
		return err
	}

	// ==========================================
	// 🚨 核弹级缓存撕毁：不仅撕毁标签，还要撕毁所有题目相关的缓存！
	// 因为题目关联的标签已经没了，必须强制刷新题目列表和详情
	// ==========================================

	// 1. 撕毁标签自身缓存
	cache.Rdb.Del(ctx, TagCacheKey)

	// 2. 撕毁所有题目列表和详情缓存 (模糊匹配)
	keys1, _ := cache.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	keys2, _ := cache.Rdb.Keys(ctx, "cache:problem:detail:*").Result()

	allKeys := append(keys1, keys2...)
	if len(allKeys) > 0 {
		cache.Rdb.Del(ctx, allKeys...)
		fmt.Printf("🗑️ 标签 %s 被删除，已强制清空全站题库相关缓存！\n", tagID)
	}

	return nil
}
