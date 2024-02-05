package dao

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type GORMArticleDAO struct {
	db *gorm.DB
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{db: db}
}

func (dao *GORMArticleDAO) Upsert(ctx context.Context, art PublicArticle) error {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.Clauses(clause.OnConflict{
		// MySql 只需要这个字段
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"utime":   now,
		}),
	}).Create(&art).Error
	// MySQL生成的语句: INSERT xxx ON DUPLICATE KEY UPDATE xxx
	// MySQL 的 upsert 语句不支持查询条件
	return err
}

func (dao *GORMArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	// 在DAO层面控制事务，不能跨库，因此操作的是两个不同的表
	// 闭包形态的事务，由GORM负责管理事务的生命周期
	var id = art.Id
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		// 业务逻辑
		var (
			err error
		)
		// 操作制作库
		// 由于MySQL的upsert语句不支持where语句，因此只能分开为更新和插入语句
		txDAO := NewGORMArticleDAO(tx)
		if id > 0 {
			err = txDAO.UpdateById(ctx, art)
		} else {
			id, err = txDAO.Insert(ctx, art)
		}
		if err != nil {
			return err
		}
		// 操作线上库
		return txDAO.Upsert(ctx, PublicArticle{Article: art})
	})
	return id, err
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
