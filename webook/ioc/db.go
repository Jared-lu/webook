package ioc

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"webook/webook/internal/repository/dao"
)

func InitDB() *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var c Config
	// 这里不要有多级路径
	err := viper.UnmarshalKey("db", &c)
	if err != nil {
		fmt.Println("初始化数据库配置失败")
	}
	db, err := gorm.Open(mysql.Open(c.DSN))
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
