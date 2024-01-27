package dao

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"time"
)

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{db: db}
}

func (dao *GORMArticleDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.Id, err
}

func (dao *GORMArticleDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	art.Utime = now
	// 不要依赖 gorm 忽略零值更新的特性，会用主键进行更新
	// 这样可读性很差
	//err := dao.db.WithContext(ctx).Create(&art).Error
	res := dao.db.WithContext(ctx).Model(&art).
		Where("id = ? AND author_id = ?", art.Id, art.AuthorId).Updates(map[string]any{
		"title":   art.Title,
		"content": art.Content,
		"utime":   art.Utime,
	})
	if res.RowsAffected == 0 {
		return fmt.Errorf("更新失败，可能是用户非法: id = %d, authorId = %d", art.Id, art.AuthorId)
	}
	return res.Error
}
