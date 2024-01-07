package repository

import (
	"context"
	"database/sql"
	"time"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
)

var (
	ErrUserDuplicate = dao.ErrUserDuplicate
	ErrUserNotFound  = dao.ErrUserNotFound
)

type UserRepository struct {
	dao   *dao.UserDAO
	cache *cache.RedisUserCache
}

func NewUserRepository(dao *dao.UserDAO, cache *cache.RedisUserCache) *UserRepository {
	return &UserRepository{dao: dao, cache: cache}
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *UserRepository) entityToDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Password: user.Password,
		Phone:    user.Phone.String,
		UserInfo: domain.UserInfo{
			NickName:    user.NickName,
			Birthday:    user.Birthday,
			Description: user.Description,
		},
		Ctime: time.UnixMilli(user.Ctime),
	}
}

func (r *UserRepository) domainToEntity(user domain.User) dao.User {
	return dao.User{
		Id: user.Id,
		Email: sql.NullString{
			String: user.Email,
			Valid:  user.Email != "",
		},
		Phone: sql.NullString{
			String: user.Phone,
			Valid:  user.Phone != "",
		},
		Password:    user.Password,
		NickName:    user.NickName,
		Birthday:    user.Birthday,
		Description: user.Description,

		Ctime: user.Ctime.UnixMilli(),
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	user, err := r.cache.Get(ctx, id)
	if err == nil {
		return user, nil
	}
	// 这里就是如果Redis没有数据，或者是崩溃了，仍然要从数据库中查
	u, err := r.dao.FindById(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	user = r.entityToDomain(u)
	go func() {
		r.cache.Set(ctx, user)
	}()
	return user, nil
}
