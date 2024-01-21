package domain

import "time"

// User 用户领域对象，业务模型
type User struct {
	Id         int64 // 用户唯一Id，由数据库生成
	Email      string
	Password   string
	Phone      string
	WechatInfo WechatInfo
	UserInfo
	Ctime time.Time
}

type UserInfo struct {
	NickName string
	// 年-月-日，如1999-01-11
	Birthday string
	// 个人简介
	Description string
}
