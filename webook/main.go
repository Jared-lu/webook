package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/cache"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
	"webook/webook/internal/web/middleware"
	"webook/webook/pkg/ginx/middlewares/ratelimit"
)

func main() {
	db := initDB()
	redis := initRedis()
	server := initWebServer()
	u := initUser(db, redis)
	u.RegisterRouter(server)
	server.Run(":8080")
}

func initUser(db *gorm.DB, redis redis.Cmdable) *web.UserHandler {
	dao := dao.NewUserDAO(db)
	userCache := cache.NewRedisUserCache(redis, time.Minute*30)
	repo := repository.NewUserRepository(dao, userCache)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	return u
}

func initDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	if err != nil {
		panic(err)
	}
	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}

func initRedis() redis.Cmdable {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	return redisClient
}

func initWebServer() *gin.Engine {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	server := gin.Default()
	server.Use(cordHdl(),
		//session(),
		middleware.NewLoginJWTMiddleWareBuilder().
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login").Build(),
		// 限流，每秒100个请求
		ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	)

	return server
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
