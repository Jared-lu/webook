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

func (dao *GORMArticleDAO) ListPub(ctx context.Context, start time.Time, offset int, limit int) ([]Article, error) {
	var res []Article
	err := dao.db.WithContext(ctx).
		Where("utime<?", start.UnixMilli()).
		Order("utime DESC").Offset(offset).Limit(limit).Find(&res).Error
	return res, err
}

func (dao *GORMArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var pub PublishedArticle
	err := dao.db.WithContext(ctx).
		Where("id = ?", id).
		First(&pub).Error
	return pub, err
}

func (dao *GORMArticleDAO) GetByAuthor(ctx context.Context, authorId int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	// SELECT * FROM XXX WHERE XX order by aaa
	// 在设计 order by 语句的时候，要注意让 order by 中的数据命中索引
	// SQL 优化的案例：早期的时候，
	// 我们的 order by 没有命中索引的，内存排序非常慢
	// 你的工作就是优化了这个查询，加进去了索引
	// author_id => author_id, utime 的联合索引
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", authorId).
		Offset(offset).
		Limit(limit).
		// 升序排序。 utime ASC
		// 混合排序
		// ctime ASC, utime desc
		Order("utime DESC").
		//Order(clause.OrderBy{Columns: []clause.OrderByColumn{
		//	{Column: clause.Column{Name: "utime"}, Desc: true},
		//	{Column: clause.Column{Name: "ctime"}, Desc: false},
		//}}).
		Find(&arts).Error
	return arts, err
}

func (dao *GORMArticleDAO) GetById(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("id = ?", id).
		First(&art).Error
	return art, err
}

func NewGORMArticleDAO(db *gorm.DB) ArticleDAO {
	return &GORMArticleDAO{db: db}
}

func (dao *GORMArticleDAO) SyncStatus(ctx context.Context, id int64, authorId int64, status uint8) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UnixMilli()
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", id, authorId).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			})
		if res.Error != nil {
			// 数据库有问题
			return res.Error
		}
		if res.RowsAffected != 1 {
			// 要么id是错的，要么作者不对
			// 这可能是由黑客入侵
			return fmt.Errorf("非法操作 uid: %d, aid: %d", authorId, id)
		}
		return tx.Model(&PublishedArticle{}).Where("id = ?", id).
			Updates(map[string]any{
				"status": status,
				"utime":  now,
			}).Error
	})
}

func (dao *GORMArticleDAO) Upsert(ctx context.Context, art PublishedArticle) error {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	err := dao.db.Clauses(clause.OnConflict{
		// MySql 只需要这个字段
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
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
	//var id = art.Id
	//err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
	//	// 业务逻辑
	//	var (
	//		err error
	//	)
	//	// 操作制作库
	//	// 由于MySQL的upsert语句不支持where语句，因此只能分开为更新和插入语句
	//	txDAO := NewGORMArticleDAO(tx)
	//	if id > 0 {
	//		err = txDAO.UpdateById(ctx, art)
	//	} else {
	//		id, err = txDAO.Insert(ctx, art)
	//	}
	//	if err != nil {
	//		return err
	//	}
	//	// 操作线上库
	//	return txDAO.Upsert(ctx, PublishedArticle{Article: art})
	//})
	//return id, err

	tx := dao.db.WithContext(ctx).Begin()
	now := time.Now().UnixMilli()
	defer tx.Rollback()
	txDAO := NewGORMArticleDAO(tx)
	var (
		id  = art.Id
		err error
	)
	if id == 0 {
		id, err = txDAO.Insert(ctx, art)
	} else {
		err = txDAO.UpdateById(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.Id = id
	publishArt := PublishedArticle{
		Article: art,
	}
	publishArt.Utime = now
	publishArt.Ctime = now
	err = tx.Clauses(clause.OnConflict{
		// ID 冲突的时候。实际上，在 MYSQL 里面你写不写都可以
		Columns: []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   art.Title,
			"content": art.Content,
			"status":  art.Status,
			"utime":   now,
		}),
	}).Create(&publishArt).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, tx.Error
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
		"status":  art.Status,
		"utime":   art.Utime,
	})
	if res.RowsAffected == 0 {
		return fmt.Errorf("更新失败，可能是用户非法: id = %d, authorId = %d", art.Id, art.AuthorId)
	}
	return res.Error
}
