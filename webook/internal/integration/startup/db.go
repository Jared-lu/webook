package startup

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
	err = initTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initTable(db *gorm.DB) error {
	// gorm自动建表
	return db.AutoMigrate(&dao.User{})
}
