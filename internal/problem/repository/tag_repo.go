package repository

import (
	"context"
	"gojo/infrastructure/mysql"
	"gojo/internal/problem/model"

	"gorm.io/gorm"
)

type TagRepository interface {
	GetTagList(ctx context.Context, tags *[]model.Tag) error
	CreateTag(ctx context.Context, tag *model.Tag) error
	DeleteTag(ctx context.Context, tagId string) error
}

type TagRepositoryMysql struct{}

func NewTagRepository() TagRepository {
	return &TagRepositoryMysql{}
}
func (r *TagRepositoryMysql) GetTagList(ctx context.Context, tags *[]model.Tag) error {
	// 2. Redis 没有，去 MySQL 查全量
	return mysql.DB.WithContext(ctx).Find(tags).Error
}

func (r *TagRepositoryMysql) CreateTag(ctx context.Context, tag *model.Tag) error {
	return mysql.DB.WithContext(ctx).Create(&tag).Error
}

func (r *TagRepositoryMysql) DeleteTag(ctx context.Context, tagID string) error {
	return mysql.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var tag model.Tag
		if err := tx.First(&tag, tagID).Error; err != nil {
			return err // 给 Controller 报 404
		}
		// 1. 极其安全：先去中间表把包含这个 Tag 的桥梁全部炸毁
		if err := tx.Model(&tag).Association("Problems").Clear(); err != nil {
			return err
		}
		// 2. 然后：把孤立无援的 Tag 本身删掉
		if err := tx.Delete(&tag).Error; err != nil {
			return err
		}
		return nil // 完美执行，提交事务！
	})
}
