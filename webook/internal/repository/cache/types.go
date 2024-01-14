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

// Cache 统一缓存API
//type Cache interface {
//	Get(ctx context.Context, key string) (any, error)
//	Set(ctx context.Context, key string, val any, expiration time.Duration)error
//}
