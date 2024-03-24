package dao

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-sql-driver/mysql"
	"github.com/olivere/elastic/v7"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

const UserIndexName = "user_index"

// 错误码定义
var (
	ErrUserDuplicate = errors.New("邮箱冲突")
	ErrUserNotFound  = gorm.ErrRecordNotFound // 不用显示返回这个Err，上层会自动匹配
)

type GormUserDAO struct {
	db     *gorm.DB
	client *elastic.Client
}

func (dao *GormUserDAO) Search(ctx context.Context, keywords []string) ([]User, error) {
	// 假定上面传入的 keywords 是经过了处理的
	queryString := strings.Join(keywords, " ")
	query := elastic.NewBoolQuery().Must(elastic.NewMatchQuery("nickname", queryString))
	resp, err := dao.client.Search(UserIndexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]User, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		var ele User
		err = json.Unmarshal(hit.Source, &ele)
		if err != nil {
			return nil, err
		}
		res = append(res, ele)
	}
	return res, nil
}

func (dao *GormUserDAO) InputUser(ctx context.Context, u User) error {
	_, err := dao.client.Index().
		Index(UserIndexName).
		Id(strconv.FormatInt(u.Id, 10)).
		BodyJson(u).Do(ctx)
	return err
}

func NewUserDAO(db *gorm.DB) UserDAO {
	return &GormUserDAO{db: db}
}

func (dao *GormUserDAO) FindByWechatOpenId(ctx context.Context, openId string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("wechat_open_id = ?", openId).First(&u).Error
	return u, err
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
