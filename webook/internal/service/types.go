package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"webook/webook/internal/domain"
)

//go:generate mockgen -source=./types.go -package=svcmocks -destination=./mocks/service.mock.go

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, user domain.User) (domain.User, error)
	Edit(ctx context.Context, user domain.User) error
	FindOrCreateByPhone(ctx context.Context, phone string) (domain.User, error)
	Profile(ctx context.Context, user domain.User) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error)
}

// CodeService 验证码服务
type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type ArticleService interface {
	// Save 保存文章，并返回文章ID
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	PublishV1(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, art domain.Article) error
	List(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPublishedById(ctx *gin.Context, id int64, uid int64) (domain.Article, error)
}
