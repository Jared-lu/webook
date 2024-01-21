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

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (svc *userService) FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error) {
	u, err := svc.repo.FindByWechatOpenId(ctx, info.OpenId)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, err
	}
	// 没有这个用户
	u = domain.User{
		WechatInfo: info,
	}
	// 创建用户
	err = svc.repo.Create(ctx, u)
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return u, err
	}
	// 这里会遇到主从延迟的问题
	return svc.repo.FindByWechatOpenId(ctx, u.WechatInfo.OpenId)
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

func (svc *userService) Edit(ctx context.Context, user domain.User) error {
	u, err := svc.repo.FindById(ctx, user.Id)
	if errors.Is(err, repository.ErrUserNotFound) {
		return ErrInvalidEmailOrPassword
	}
	if err != nil {
		return err
	}

	// 更新用户的信息
	u.NickName = user.NickName
	u.Birthday = user.Birthday
	u.Description = user.Description
	err = svc.repo.Update(ctx, u)
	return err
}

func (svc *userService) FindOrCreateByPhone(ctx context.Context, phone string) (domain.User, error) {
	// 请求来到这里，一部分人是新用户注册，一部分是登录
	// 对于登录的，只需要向数据库发起读请求，只剩下注册的需要发起写请求
	// 从而避免了大量的写请求打到数据库上
	// 对于登录的人来说，读要比写更快
	// 因此不要把先查的部分去掉，虽然注册的用户需要发起两次数据库请求
	// 分开两个路径还有一个好处就是，在系统降级时只走快路径，优先服务已注册用户
	// 快路径，对登录的人
	u, err := svc.repo.FindByPhone(ctx, phone)
	//要判断有没有这个用户
	if !errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, err
	}
	// 没有这个用户
	u = domain.User{
		Phone: phone,
	}
	// 创建用户
	// 慢路径
	err = svc.repo.Create(ctx, u)
	// 这样解决主从延迟问题导致的FindByPhone找不到用户
	//  return svc.repo.CreateV1(ctx, u)
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return u, err
	}
	// 这里会遇到主从延迟的问题
	// 因为往主库写入一个新用户了，但这个查询用户的请求发到从库去读的时候可能就没有这条记录
	// 唯一解决的方法就是DAO的Insert接口要返回需要的数据或者整个user
	// 因为要返回一个user id给 web 层
	return svc.repo.FindByPhone(ctx, u.Phone)
}

func (svc *userService) Profile(ctx context.Context, user domain.User) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, user.Id)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidEmailOrPassword
	}
	return u, err
}
