package dao

import "database/sql"

// User 用户数据库模型
type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 设置为唯一索引
	Email    sql.NullString `gorm:"unique"`
	Password string

	// 唯一索引允许有多个为NULL，但不允许有多个空字符串""
	Phone sql.NullString `gorm:"unique"`

	// 微信字段
	WechatUnionId sql.NullString `gorm:"unique"`
	WechatOpenId  sql.NullString `gorm:"unique"`

	// 其他用户个人信息
	NickName    string
	Birthday    string
	Description string

	// 创建时间 毫秒数
	Ctime int64
	// 更新时间 毫秒数
	Utime int64
}

type Article struct {
	Id    int64  `gorm:"primaryKey,autoIncrement"`
	Title string `gorm:"type=varchar(1024)"`
	// 对于关系型数据库用
	Content string `gorm:"type=BLOB"`
	// 如何设计索引
	// 最常用的就是在 WHERE 的字段上创建
	// 在帖子这里，查询场景是什么样的？
	// - 对于一个创作者，查看自己的所有文章
	//   产品经理说要根据创建时间倒序排序
	// 	以下这个语句使用 author_id 和 ctime 联合索引是最合适的，因为索引天然有序，数据库不需要再进行一次排序
	// 	SELECT * FROM article WHERE author_id = xxx ORDER BY `ctime` DESC;
	// - 单独查询某一篇文章，id 是主键，天然命中索引
	//	SELECT * FROM articles WHERE id = xxx;

	// 区别以下两种建立索引的区别：
	// 1. 在 author_id和ctime上创建联合索引
	//AuthorId int64 `gorm:"index=aid_ctime"`
	//Ctime    int64 `gorm:"index=aid_ctime"`

	// 2. 在 author_id上创建索引
	AuthorId int64 `gorm:"index"`
	Ctime    int64
	Utime    int64
}
