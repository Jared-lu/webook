package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

// CodeCache 验证码缓存
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type RedisUserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

// Cache 统一缓存API
//type Cache interface {
//	Get(ctx context.Context, key string) (any, error)
//	Set(ctx context.Context, key string, val any, expiration time.Duration)error
//}
