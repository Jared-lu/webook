package startup

import (
	"context"
	"database/sql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
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

var db *gorm.DB

// InitTestDB 测试的话，不用控制并发。等遇到了并发问题再说
func InitTestDB() *gorm.DB {
	if db == nil {
		dsn := "root:root@tcp(localhost:13316)/webook"
		sqlDB, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}
		for {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err = sqlDB.PingContext(ctx)
			cancel()
			if err == nil {
				break
			}
			log.Println("等待连接 MySQL", err)
		}
		db, err = gorm.Open(mysql.Open(dsn))
		if err != nil {
			panic(err)
		}
		err = initTable(db)
		if err != nil {
			panic(err)
		}
		db = db.Debug()
	}
	return db
}

func initTable(db *gorm.DB) error {
	// gorm自动建表
	return db.AutoMigrate(&dao.User{}, &dao.Article{})
}
