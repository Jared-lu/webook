package cache

import (
	"context"
	"errors"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"sync"
	"time"
)

var (
	ErrCodeSendTooMany   = errors.New("验证码发送太频繁")
	ErrCodeVerifyTooMany = errors.New("验证次数太多")
	ErrUnknown           = errors.New("未知错误")
	ErrKeyNotExist       = errors.New("key不存在")
)

type CodeLocalCache struct {
	cache      *lru.Cache
	lock       sync.Mutex
	expiration time.Duration
}

func NewCodeLocalCache(cache *lru.Cache, expiration time.Duration) *CodeLocalCache {
	return &CodeLocalCache{cache: cache, expiration: expiration}
}

func (c *CodeLocalCache) Set(ctx context.Context, biz, phone, code string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := c.key(biz, phone)
	now := time.Now()
	val, ok := c.cache.Get(key)
	if !ok {
		// key不存在
		c.cache.Add(key, codeItem{
			code: code,
			cnt:  3,
			// 当前时间+key过期时间
			expire: now.Add(c.expiration),
		})
		return nil
	}
	item, ok := val.(codeItem)
	if !ok {
		return errors.New("系统错误")
	}
	// key的expire - 当前时间
	if item.expire.Sub(now) > time.Minute*9 {
		// 不到一分钟
		return ErrCodeSendTooMany
	}
	// 重发
	c.cache.Add(key, codeItem{
		code:   code,
		cnt:    3,
		expire: now.Add(c.expiration),
	})
	return nil
}

func (c *CodeLocalCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	key := c.key(biz, phone)
	val, ok := c.cache.Get(key)
	if !ok {
		// 没发验证码
		return false, ErrKeyNotExist
	}
	item, ok := val.(codeItem)
	if !ok {
		return false, errors.New("系统错误")
	}
	if item.cnt <= 0 {
		return false, ErrCodeVerifyTooMany
	}
	item.cnt--
	return item.code == inputCode, nil
}

func (c *CodeLocalCache) key(biz, phone string) string {
	// 验证码的key命名方式实例: phone_code:login:13711112222
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

type codeItem struct {
	code string
	// 可验证次数
	cnt int
	// 过期时间，需要自己维持
	expire time.Time
}
