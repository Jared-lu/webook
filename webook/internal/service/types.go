package service

import (
	"context"
	"webook/webook/internal/domain"
)

//go:generate mockgen -source=./types.go -package=svcmocks -destination=./mocks/service.mock.go

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, user domain.User) (domain.User, error)
	Edit(ctx context.Context, user domain.User) error
	FindOrCreateByPhone(ctx context.Context, phone string) (domain.User, error)
	Profile(ctx context.Context, user domain.User) (domain.User, error)
}

// CodeService 验证码服务
type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}
