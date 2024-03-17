package ioc

import (
	"fmt"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/prometheus"
	"webook/webook/internal/repository/dao"
	"webook/webook/pkg/logger"
)

func InitDB(l logger.Logger) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var c Config
	// 这里不要有多级路径
	err := viper.UnmarshalKey("db", &c)
	if err != nil {
		fmt.Println("初始化数据库配置失败")
	}
	db, err := gorm.Open(mysql.Open(c.DSN), &gorm.Config{
		Logger: glogger.New(gormLogger(l.Debug), glogger.Config{
			SlowThreshold: 0,
			LogLevel:      glogger.Info,
		})})
	if err != nil {
		panic(err)
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		StartServer:     false,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	err = initTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

type gormLogger func(msg string, args ...logger.Field)

// Printf 适配器模式
// 但会引起复制
func (g gormLogger) Printf(msg string, args ...interface{}) {
	g(msg, logger.Field{
		Key:   "args",
		Value: args,
	})
}

func initTable(db *gorm.DB) error {
	// gorm自动建表
	return db.AutoMigrate(&dao.User{})
}
