package ratelimit

import "context"

//go:generate mockgen -source=./types.go -package=ratelimitmocks -destination=./mocks/ratelimit.mock.go

// Limiter 限流器
type Limiter interface {
	// Limit 限流方法
	// key 限流对象
	// 返回值：
	// bool 表示是否限流，true 为触发限流
	// error表示限流器是否出现错误
	Limit(ctx context.Context, key string) (bool, error)
}
