package dao

import "database/sql"

// User 用户数据库模型
type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 设置为唯一索引
	Email    sql.NullString `gorm:"unique"`
	Password string

	// 唯一索引允许为null，但不允许为""
	Phone sql.NullString `gorm:"unique"`

	// 其他用户个人信息
	NickName    string
	Birthday    string
	Description string

	// 创建时间 毫秒数
	Ctime int64
	// 更新时间 毫秒数
	Utime int64
}