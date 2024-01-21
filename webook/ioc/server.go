package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webook/webook/internal/web"
	"webook/webook/internal/web/middleware"
	"webook/webook/pkg/ginx/middlewares/ratelimit"
	ratelimit2 "webook/webook/pkg/ginx/ratelimit"
)

func InitGinServer(middlewares []gin.HandlerFunc, userHandler *web.UserHandler,
	wechatHandler *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(middlewares...)
	// 注册路由
	userHandler.RegisterRouter(server)
	wechatHandler.RegisterRoutes(server)
	return server
}

// initLimiterOfAccess 服务端的访问限流器
func initLimiterOfAccess(cmd redis.Cmdable) ratelimit2.Limiter {
	// 每秒限流100个请求
	return ratelimit2.NewRedisSlidingWindowLimiter(cmd, time.Second, 100)
}

func InitGinMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cordHdl(),
		middleware.NewLoginJWTMiddleWareBuilder().
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login").
			IgnorePaths("/users/login_sms/code/send").
			IgnorePaths("/users/login_sms").
			IgnorePaths("/oauth2/wechat/oauth2url").
			IgnorePaths("/oauth2/wechat/callback").Build(),
		ratelimit.NewBuilder(initLimiterOfAccess(redisClient)).Build(),
	}
}

// cordHdl 跨域请求
func cordHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		// 跨域允许接受的来源
		AllowOrigins: []string{"http://localhost:3000"},
		// 跨域允许接受的方法
		AllowMethods: []string{"PUT", "PATCH", "POST", "GET"},
		// 跨域允许接受的首部
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 允许前端拿到服务器返回的Header，JWT会用到
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
		// 是否允许带 cookie 之类的东西
		AllowCredentials: true,
		// 与 AllowOrigins 作用一样，当功能更强大
		// origin 是Preflight请求首部 Origin
		AllowOriginFunc: func(origin string) bool {
			if strings.Contains(origin, "localhost") {
				return true
			}
			return strings.Contains(origin, "yourcompany.com")
		},
		// Preflight请求过期时间
		MaxAge: 12 * time.Hour,
	})
}

func session() gin.HandlerFunc {
	// 设置Session
	// 指定Session的数据存储的地方为cookie
	//store := cookie.NewStore([]byte("secret"))
	//server.Use(sessions.Sessions("MySession", store)) // 这个MySession肯定是放在Cookie中，它带有sess_id，不管Session存储在哪里

	store := memstore.NewStore([]byte("HiIilLa4O8Xy3Pm8C5mh5HymYaYt9eTj"),
		[]byte("bNzvaKyNGy76lGs3BSaoY3fn9ketFdDf"))
	return sessions.Sessions("mysession", store)
}
