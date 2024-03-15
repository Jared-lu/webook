package dao

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var ErrWaitingSMSNotFound = gorm.ErrRecordNotFound

const (
	// 等待发送
	asyncStatusWaiting = iota
	// 失败了，并且超过了重试次数
	asyncStatusFailed
	// 发送成功
	asyncStatusSuccess
)

type GORMAsyncSmsDAO struct {
	db *gorm.DB
}

func NewGORMAsyncSmsDAO(db *gorm.DB) AsyncSmsDAO {
	return &GORMAsyncSmsDAO{
		db: db,
	}
}

func (g *GORMAsyncSmsDAO) Insert(ctx context.Context, s AsyncSms) error {
	return g.db.Create(&s).Error
}

// GetWaitingSMS 查找等待异步发送的短信请求
func (g *GORMAsyncSmsDAO) GetWaitingSMS(ctx context.Context) (AsyncSms, error) {
	var s AsyncSms
	err := g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 只找 1 分钟前的异步短信发送
		now := time.Now().UnixMilli()
		endTime := now - time.Minute.Milliseconds()

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("utime < ? and status = ?",
				endTime, asyncStatusWaiting).First(&s).Error
		if err != nil {
			return err
		}

		err = tx.Model(&AsyncSms{}).
			Where("id = ?", s.Id).
			Updates(map[string]any{
				"retry_cnt": gorm.Expr("retry_cnt + 1"),
				"utime":     now,
			}).Error
		return err
	})
	return s, err
}

func (g *GORMAsyncSmsDAO) MarkSuccess(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&AsyncSms{}).
		Where("id =?", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusSuccess,
		}).Error
}

func (g *GORMAsyncSmsDAO) MarkFailed(ctx context.Context, id int64) error {
	now := time.Now().UnixMilli()
	return g.db.WithContext(ctx).Model(&AsyncSms{}).
		// 只有到达了重试次数才会更新
		Where("id =? and `retry_cnt`>=`retry_max`", id).
		Updates(map[string]any{
			"utime":  now,
			"status": asyncStatusFailed,
		}).Error
}

type AsyncSms struct {
	Id     int64
	Config sqlx.JsonColumn[SmsConfig]
	// 重试次数
	RetryCnt int
	// 重试的最大次数
	RetryMax int
	Status   uint8
	Ctime    int64
	Utime    int64 `gorm:"index"`
}

type SmsConfig struct {
	TplId   string
	Args    []string
	Numbers []string
}
