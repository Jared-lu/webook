package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"time"
	"webook/webook/internal/repository"
	"webook/webook/internal/repository/dao"
	"webook/webook/internal/service"
	"webook/webook/internal/web"
)

func main() {
	server := initWebServer()
	db := initDB()
	dao := dao.NewUserDAO(db)
	repo := repository.NewUserRepository(dao)
	svc := service.NewUserService(repo)
	u := web.NewUserHandler(svc)
	u.RegisterRouter(server)
	server.Run(":8080")
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

func initWebServer() *gin.Engine {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		// 跨域允许接受的来源
		AllowOrigins: []string{"http://localhost:3000"},
		// 跨域允许接受的方法
		AllowMethods: []string{"PUT", "PATCH", "POST", "GET"},
		// 跨域允许接受的首部
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 允许前端拿到返回的Header，JWT会用到
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
	}))
	return server
}
