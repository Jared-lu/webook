package ioc

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webook/webook/internal/repository/dao"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
