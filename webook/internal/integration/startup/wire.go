//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
)

func InitApp() *gin.Engine {
	wire.Build(
		/******** 最底层依赖 ********/
		InitDB, InitRedis, InitSMSService,
		dao.NewUserDAO,
		cache.NewRedisUserCache, cache.NewRedisCodeCache,
		repository.NewUserRepository, repository.NewCacheCodeRepository,
		service.NewUserService, service.NewSmsCodeService,
		web.NewUserHandler,
		InitGinServer, InitGinMiddlewares,
	)
	return new(gin.Engine)
}
