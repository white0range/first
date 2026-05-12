// services/tag_service.go
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"gojo/global"
	"gojo/models"
	"time"
)

const TagCacheKey = "cache:tags:all" // 全站标签只有一份，固定 Key

// GetTagList 获取全站标签（走缓存）
func GetTagList(ctx context.Context) ([]models.Tag, error) {
	var tags []models.Tag

	// 1. 先问 Redis 要
	cachedData, err := global.Rdb.Get(ctx, TagCacheKey).Result()
	if err == nil {
		fmt.Println("⚡ 触发 Redis 缓存！全站标签极速返回！")
		json.Unmarshal([]byte(cachedData), &tags)
		return tags, nil
	}

	// 2. Redis 没有，去 MySQL 查全量
	if err := models.DB.Find(&tags).Error; err != nil {
		return nil, err
	}

	// 3. 查完塞给 Redis (标签很久才变一次，直接给个 7 天过期时间)
	jsonBytes, _ := json.Marshal(tags)
	global.Rdb.Set(ctx, TagCacheKey, jsonBytes, 7*24*time.Hour)

	return tags, nil
}

// CreateTag 创建标签并撕毁标签缓存
func CreateTag(ctx context.Context, name string) (*models.Tag, error) {
	tag := models.Tag{Name: name}

	// 插入数据库 (如果 name 有 Unique 索引，重复会报错)
	if err := models.DB.Create(&tag).Error; err != nil {
		return nil, err
	}

	// 🚨 撕毁标签缓存，下次获取就会拿最新版
	global.Rdb.Del(ctx, TagCacheKey)

	return &tag, nil
}

// DeleteTag 删除标签，级联清理中间表，并触发“核弹级”缓存撕毁
func DeleteTag(ctx context.Context, tagID string) error {
	var tag models.Tag
	if err := models.DB.First(&tag, tagID).Error; err != nil {
		return err // 给 Controller 报 404
	}

	// 💥 GORM 魔法一：级联删除 (清理 tag 本身 + problem_tags 中间表)
	if err := models.DB.Select("Problems").Delete(&tag).Error; err != nil {
		return err
	}

	// ==========================================
	// 🚨 核弹级缓存撕毁：不仅撕毁标签，还要撕毁所有题目相关的缓存！
	// 因为题目关联的标签已经没了，必须强制刷新题目列表和详情
	// ==========================================

	// 1. 撕毁标签自身缓存
	global.Rdb.Del(ctx, TagCacheKey)

	// 2. 撕毁所有题目列表和详情缓存 (模糊匹配)
	keys1, _ := global.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	keys2, _ := global.Rdb.Keys(ctx, "cache:problem:detail:*").Result()

	allKeys := append(keys1, keys2...)
	if len(allKeys) > 0 {
		global.Rdb.Del(ctx, allKeys...)
		fmt.Printf("🗑️ 标签 %s 被删除，已强制清空全站题库相关缓存！\n", tagID)
	}

	return nil
}
