//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/webook/internal/repository"
	cache "webook/webook/internal/repository/cache/Redis"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	web2 "webook/webook/internal/web/jwt"
	"webook/webook/ioc"
)

func initApp() *gin.Engine {
	wire.Build(
		/******** 最底层依赖 ********/
		ioc.InitDB, ioc.InitRedis, ioc.InitSMSService,
		dao.NewUserDAO,
		cache.NewRedisUserCache, cache.NewRedisCodeCache,
		repository.NewUserRepository, repository.NewCacheCodeRepository,
		service.NewUserService, service.NewSmsCodeService,
		web.NewUserHandler, web.NewOAuth2WechatHandler, web2.NewRedisJWTHandler,
		ioc.InitGinServer, ioc.InitGinMiddlewares,
		ioc.InitOAuth2WechatService,
	)
	return new(gin.Engine)
}
