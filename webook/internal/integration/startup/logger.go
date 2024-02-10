package startup

import (
	"go.uber.org/zap"
	"webook/webook/pkg/logger"
)

// InitLogger 这是初始化全局的logger
func InitLogger() *zap.Logger {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return l
}

func InitZapLogger() logger.Logger {
	// 这里默认就用控制台作为输出
	// 生产环境要用文件
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
