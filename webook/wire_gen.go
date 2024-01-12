// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	"webook/webook/ioc"
)

// Injectors from wire.go:

func initApp() *gin.Engine {
	cmdable := ioc.InitRedis()
	v := ioc.InitGinMiddlewares(cmdable)
	db := ioc.InitDB()
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCacheCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewSmsCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService)
	engine := ioc.InitGinServer(v, userHandler)
	return engine
}
