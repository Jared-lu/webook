//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/webook/internal/repository"
	cache "webook/webook/internal/repository/cache/Redis"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	web2 "webook/webook/internal/web/jwt"
)

func InitApp() *gin.Engine {
	wire.Build(
		/******** 最底层依赖 ********/
		InitDB, InitRedis,
		dao.NewUserDAO,
		cache.NewRedisUserCache, cache.NewRedisCodeCache,
		repository.NewUserRepository, repository.NewCacheCodeRepository,
		service.NewUserService, service.NewSmsCodeService,
		InitOAuth2WechatService, InitSMSService,
		web.NewUserHandler, web.NewOAuth2WechatHandler, web2.NewRedisJWTHandler,
		/******** 公共组件 ********/
		InitZapLogger, InitGinMiddlewares,
		/******** 初始化Server ********/
		InitGinServer,
	)
	return new(gin.Engine)
}

var thirdProvider = wire.NewSet(InitDB, InitRedis, InitZapLogger)

// InitArticleHandler 单独初始化某一部分，可以更好的为测试而定制
func InitArticleHandler() *web.ArticleHandler {
	wire.Build(web.NewArticleHandler, service.NewArticleService, repository.NewCacheArticleRepository,
		dao.NewGORMArticleDAO,
		thirdProvider)
	return new(web.ArticleHandler)
}
