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

type CacheUserRepository struct {
	dao   dao.UserDAO
	cache cache.UserCache
}

func NewUserRepository(dao dao.UserDAO, cache cache.UserCache) UserRepository {
	return &CacheUserRepository{dao: dao, cache: cache}
}

func (r *CacheUserRepository) FindByWechatOpenId(ctx context.Context, OpenId string) (domain.User, error) {
	u, err := r.dao.FindByWechatOpenId(ctx, OpenId)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}

func (r *CacheUserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, r.domainToEntity(u))
}

func (r *CacheUserRepository) CreateV1(ctx context.Context, u domain.User) (domain.User, error) {
	user, err := r.dao.InsertV1(ctx, r.domainToEntity(u))
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *CacheUserRepository) entityToDomain(user dao.User) domain.User {
	return domain.User{
		Id:       user.Id,
		Email:    user.Email.String,
		Password: user.Password,
		Phone:    user.Phone.String,
		WechatInfo: domain.WechatInfo{
			OpenId:  user.WechatOpenId.String,
			UnionId: user.WechatUnionId.String,
		},
		UserInfo: domain.UserInfo{
			NickName:    user.NickName,
			Birthday:    user.Birthday,
			Description: user.Description,
		},
		Ctime: time.UnixMilli(user.Ctime),
	}
}

func (r *CacheUserRepository) domainToEntity(user domain.User) dao.User {
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
		WechatOpenId: sql.NullString{
			String: user.WechatInfo.OpenId,
			Valid:  user.WechatInfo.OpenId != "",
		},
		WechatUnionId: sql.NullString{
			String: user.WechatInfo.UnionId,
			Valid:  user.WechatInfo.UnionId != "",
		},
		Ctime: user.Ctime.UnixMilli(),
	}
}

func (r *CacheUserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(user), nil
}

func (r *CacheUserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
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

func (r *CacheUserRepository) Update(ctx context.Context, user domain.User) error {
	err := r.dao.Update(ctx, r.domainToEntity(user))
	// 更新操作：先更新数据库，再写入缓存
	go func() {
		if err != nil {
			r.cache.Set(ctx, user)
		}
	}()
	return err
}

func (r *CacheUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := r.dao.FindByPhone(ctx, phone)
	if err != nil {
		return domain.User{}, err
	}
	return r.entityToDomain(u), nil
}
