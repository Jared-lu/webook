package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

// 错误码定义
var (
	ErrUserDuplicate = errors.New("邮箱冲突")
	ErrUserNotFound  = gorm.ErrRecordNotFound // 不用显示返回这个Err，上层会自动匹配
)

type GormUserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GormUserDAO{db: db}
}

func (dao *GormUserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	// 返回查找对象，这里的对象是数据库模型
	err := dao.db.WithContext(ctx).Where("`Id` = ?", id).First(&u).Error
	return u, err
}

func (dao *GormUserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *GormUserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error // 传ctx是为了保持链路的持续
	// 解析err，这一步是跟底层强耦合（即和具体的数据库耦合在一起）
	var mysqlErr *mysql.MySQLError
	ok := errors.As(err, &mysqlErr) // 类型断言
	if ok {
		// mysql唯一索引冲突错误码
		const uniqueConflictsErr = 1062
		if mysqlErr.Number == uniqueConflictsErr {
			// 邮箱冲突
			return ErrUserDuplicate
		}
	}
	return err
}

func (dao *GormUserDAO) InsertV1(ctx context.Context, u User) (User, error) {
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error // 传ctx是为了保持链路的持续
	// 解析err，这一步是跟底层强耦合（即和具体的数据库耦合在一起）
	var mysqlErr *mysql.MySQLError
	ok := errors.As(err, &mysqlErr) // 类型断言
	if ok {
		// mysql唯一索引冲突错误码
		const uniqueConflictsErr = 1062
		if mysqlErr.Number == uniqueConflictsErr {
			// 邮箱冲突
			return User{}, ErrUserDuplicate
		}
	}
	return u, err
}

func (dao *GormUserDAO) Update(ctx context.Context, u User) error {
	err := dao.db.Model(&u).WithContext(ctx).Where("`Id`=?", u.Id).
		Updates(User{Email: u.Email, Password: u.Password,
			NickName: u.NickName, Birthday: u.Birthday, Description: u.Description,
			Utime: time.Now().UnixMilli()}).Error
	return err
}

func (dao *GormUserDAO) FindByPhone(ctx context.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}
