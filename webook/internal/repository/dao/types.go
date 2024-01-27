package dao

import "context"

//go:generate mockgen -source=./types.go -package=daomocks -destination=./mocks/dao.mock.go

type UserDAO interface {
	FindById(ctx context.Context, id int64) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	Insert(ctx context.Context, u User) error
	InsertV1(ctx context.Context, u User) (User, error)
	Update(ctx context.Context, u User) error
	FindByPhone(ctx context.Context, phone string) (User, error)
	FindByWechatOpenId(ctx context.Context, openId string) (User, error)
}

type ArticleDAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
}
