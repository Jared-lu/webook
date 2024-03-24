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
	Status   uint8
	Ctime    int64
	Utime    int64
	Tags     []string `json:"tags"`
}

// PublishedArticle 代表线上库的文章
type PublishedArticle struct {
	Article
}

// 另一种写法
//type PublishedArticle Article

// 正常来说，一张主表和与它有关联关系的表会共用一个DAO，
// 所以我们就用一个 DAO 来操作

// Interactive 记录文章的点赞、阅读、计数
type Interactive struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	ReadCnt    int64
	CollectCnt int64
	LikeCnt    int64 `gorm:"index"`
	Ctime      int64
	Utime      int64
}

// UserLikeBiz 用户点赞的某个东西,
// 某个用户给某个资源点赞了
// 或者看作 某个资源有某个用户点赞了
type UserLikeBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 三个构成唯一索引
	BizId int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	Uid   int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	// 依旧是只在 DB 层面生效的状态
	// 1- 有效，0-无效。软删除的用法
	Status uint8
	Ctime  int64
	Utime  int64
}

// Collection 收藏夹
type Collection struct {
	Id   int64  `gorm:"primaryKey,autoIncrement"`
	Name string `gorm:"type=varchar(1024)"`
	Uid  int64  `gorm:""`

	Ctime int64
	Utime int64
}

// UserCollectionBiz 收藏的东西
type UserCollectionBiz struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 收藏夹 ID
	// 作为关联关系中的外键，我们这里需要索引
	Cid   int64  `gorm:"index"`
	BizId int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz   string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	// 这算是一个冗余，因为正常来说，
	// 只需要在 Collection 中维持住 Uid 就可以
	Uid   int64 `gorm:"uniqueIndex:biz_type_id_uid"`
	Ctime int64
	Utime int64
}

type FollowRelation struct {
	ID int64 `gorm:"primaryKey,autoIncrement,column:id"`

	Follower int64 `gorm:"type:int(11);not null;uniqueIndex:follower_followee"`
	Followee int64 `gorm:"type:int(11);not null;uniqueIndex:follower_followee"`

	Status uint8

	// 这里你可以根据自己的业务来增加字段，比如说
	// 关系类型，可以搞些什么普通关注，特殊关注
	// Type int64 `gorm:"column:type;type:int(11);comment:关注类型 0-普通关注"`
	// 备注
	// Note string `gorm:"column:remark;type:varchar(255);"`
	// 创建时间
	Ctime int64
	Utime int64
}

// UserRelation 另外一种设计方案，但是不要这么做
type UserRelation struct {
	ID     int64 `gorm:"primaryKey,autoIncrement,column:id"`
	Uid1   int64 `gorm:"column:uid1;type:int(11);not null;uniqueIndex:user_contact_index"`
	Uid2   int64 `gorm:"column:uid2;type:int(11);not null;uniqueIndex:user_contact_index"`
	Block  bool  // 拉黑
	Mute   bool  // 屏蔽
	Follow bool  // 关注
}

type FollowStatics struct {
	ID  int64 `gorm:"primaryKey,autoIncrement,column:id"`
	Uid int64 `gorm:"unique"`
	// 有多少粉丝
	Followers int64
	// 关注了多少人
	Followees int64

	Utime int64
	Ctime int64
}
