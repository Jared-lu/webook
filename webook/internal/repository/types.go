package repository

import (
	"context"
	"webook/webook/internal/domain"
)

//go:generate mockgen -source=./types.go -package=repomocks -destination=./mocks/repository.mock.go

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	CreateV1(ctx context.Context, u domain.User) (domain.User, error)
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindById(ctx context.Context, id int64) (domain.User, error)
	Update(ctx context.Context, user domain.User) error
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindByWechatOpenId(ctx context.Context, OpenId string) (domain.User, error)
}

type CodeRepository interface {
	Store(ctx context.Context, biz string, phone string, code string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	// SyncV1 存储并同步数据
	SyncV1(ctx context.Context, art domain.Article) (int64, error)
	SyncV2(ctx context.Context, art domain.Article) (int64, error)
	Sync(ctx context.Context, art domain.Article) (int64, error)
}

type ArticleAuthorRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
}

type ArticleReaderRepository interface {
	// Save 有就更新，没有就创建
	Save(ctx context.Context, art domain.Article) (int64, error)
}
