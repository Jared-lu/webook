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

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
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
