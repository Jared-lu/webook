package repository

import (
	"context"
	"webook/webook/internal/repository/cache"
	cache2 "webook/webook/internal/repository/cache/Redis"
)

var (
	ErrCodeSendTooMany   = cache2.ErrCodeSendTooMany
	ErrCodeVerifyTooMany = cache2.ErrCodeVerifyTooMany
)

type CacheCodeRepository struct {
	cache cache.CodeCache
}

func NewCacheCodeRepository(c cache.CodeCache) CodeRepository {
	return &CacheCodeRepository{
		cache: c,
	}
}

// Store 存储验证码
func (repo *CacheCodeRepository) Store(ctx context.Context, biz string,
	phone string, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

// Verify 校验验证码
func (repo *CacheCodeRepository) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, inputCode)
}
