package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	ErrCodeSendTooMany   = errors.New("验证码发送太频繁")
	ErrCodeVerifyTooMany = errors.New("验证次数太多")
	ErrUnknown           = errors.New("未知错误")
)

// 注入lua脚本
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type RedisCodeCache struct {
	client redis.Cmdable
}

func NewRedisCodeCache(client redis.Cmdable) CodeCache {
	return &RedisCodeCache{
		client: client,
	}
}

// Set 将验证码存入到redis
func (c *RedisCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	// 执行set_code.lua，将验证码存入到redis
	res, err := c.client.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	case 0:
		// 毫无问题
		return nil
	case -1:
		// 发送太频繁
		return ErrCodeSendTooMany
	default:
		// 系统错误
		return errors.New("系统错误")

	}
}

func (c *RedisCodeCache) key(biz, phone string) string {
	// 验证码的key命名方式实例: phone_code:login:13711112222
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

// Verify 校验验证码
// inputCode 用户输入的验证码
func (c *RedisCodeCache) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	res, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, inputCode).Int()
	if err != nil {
		return false, err
	}
	switch res {
	case 0:
		return true, nil
	case -1:
		return false, ErrCodeVerifyTooMany
	case -2:
		return false, nil
	default:
		return false, ErrUnknown
	}
}
