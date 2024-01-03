package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
)

var (
	ErrUserDuplicateEmail     = repository.ErrUserDuplicate
	ErrInvalidEmailOrPassword = errors.New("邮箱或密码不对") // 不区分用户不存在或密码错误
)

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, user domain.User) (domain.User, error)
	Edit(ctx context.Context, user domain.User) error
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	Profile(ctx context.Context, user domain.User) (domain.User, error)
}

type userService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (svc *userService) SignUp(ctx context.Context, u domain.User) error {
	// 加密用户密码
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// 保存加密的密码
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *userService) Login(ctx context.Context, u domain.User) (domain.User, error) {
	// 先找用户
	user, err := svc.repo.FindByEmail(ctx, u.Email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidEmailOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 比较密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password)) // 参数的顺序不能乱
	if err != nil {
		return domain.User{}, ErrInvalidEmailOrPassword
	}
	return user, nil
}

func (svc *userService) Edit(ctx context.Context, u domain.User) error {
	//TODO implement me
	panic("implement me")
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (svc *userService) Profile(ctx context.Context, u domain.User) (domain.User, error) {
	//TODO implement me
	panic("implement me")
}
