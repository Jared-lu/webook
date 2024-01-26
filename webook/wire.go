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
		ioc.InitDB, ioc.InitRedis,
		dao.NewUserDAO,
		cache.NewRedisUserCache, cache.NewRedisCodeCache,
		repository.NewUserRepository, repository.NewCacheCodeRepository,
		service.NewUserService, service.NewSmsCodeService,
		ioc.InitOAuth2WechatService, ioc.InitSMSService,
		web.NewUserHandler, web.NewOAuth2WechatHandler, web2.NewRedisJWTHandler,
		/******** 公共组件 ********/
		ioc.InitZapLogger, ioc.InitGinMiddlewares,
		/******** 初始化Server ********/
		ioc.InitGinServer,
	)
	return new(gin.Engine)
}
