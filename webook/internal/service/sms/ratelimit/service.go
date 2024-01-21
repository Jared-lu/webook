package ratelimit

import (
	"context"
	"fmt"
	"webook/webook/internal/service/sms"
	"webook/webook/pkg/ginx/ratelimit"
)

var errLimited = fmt.Errorf("触发了限流")

type Service struct {
	svc     sms.Service
	limiter ratelimit.Limiter
	// 限流对象，要用哪一家短信服务商，如腾讯云，则限制访问腾讯云服务的key访问次数
	key string
}

func NewService(svc sms.Service, limiter ratelimit.Limiter, key string) *Service {
	return &Service{
		svc:     svc,
		limiter: limiter,
		key:     key,
	}
}

func (s *Service) Send(ctx context.Context, templateID string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, s.key)
	if err != nil {
		return fmt.Errorf("短信服务判断是否限流出现问题, %w", err)
	}
	if limited {
		return errLimited
	}
	return s.svc.Send(ctx, templateID, args, numbers...)
}
