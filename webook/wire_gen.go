// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache/Redis"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	web2 "webook/webook/internal/web"
	"webook/webook/internal/web/jwt"
	"webook/webook/ioc"
)

// Injectors from wire.go:

func initApp() *gin.Engine {
	cmdable := ioc.InitRedis()
	logger := ioc.InitZapLogger()
	v := ioc.InitGinMiddlewares(cmdable, logger)
	db := ioc.InitDB(logger)
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewRedisUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewRedisCodeCache(cmdable)
	codeRepository := repository.NewCacheCodeRepository(codeCache)
	smsService := ioc.InitSMSService(cmdable)
	codeService := service.NewSmsCodeService(codeRepository, smsService)
	jwtHandler := web.NewRedisJWTHandler(cmdable)
	userHandler := web2.NewUserHandler(userService, codeService, jwtHandler, logger)
	wechatService := ioc.InitOAuth2WechatService()
	oAuth2WechatHandler := web2.NewOAuth2WechatHandler(wechatService, userService, jwtHandler)
	engine := ioc.InitGinServer(v, userHandler, oAuth2WechatHandler)
	return engine
}
