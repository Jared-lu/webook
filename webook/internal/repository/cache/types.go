package cache

import (
	"context"
)

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
