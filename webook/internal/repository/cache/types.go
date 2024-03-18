package cache

import (
	"context"
	"webook/webook/internal/domain"
)

//go:generate mockgen -source=./types.go -package=cachemocks -destination=./mocks/cache.mock.go

// UserCache 用户缓存
type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
}

// CodeCache 验证码缓存
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type ArticleCache interface {
	// GetFirstPage 只缓存第第一页的数据
	// 并且不缓存整个 Content
	GetFirstPage(ctx context.Context, author int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, author int64, arts []domain.Article) error
	DelFirstPage(ctx context.Context, author int64) error

	Set(ctx context.Context, art domain.Article) error
	Get(ctx context.Context, id int64) (domain.Article, error)

	// SetPub 已发表文章的缓存
	// 正常来说，创作者和读者的 Redis 集群要分开，因为读者是一个核心中的核心
	SetPub(ctx context.Context, article domain.Article) error
	DelPub(ctx context.Context, id int64) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
}

// Cache 统一缓存API
//type Cache interface {
//	Get(ctx context.Context, key string) (any, error)
//	Set(ctx context.Context, key string, val any, expiration time.Duration)error
//}
