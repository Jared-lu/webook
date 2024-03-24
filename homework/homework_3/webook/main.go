package webook

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	server := gin.Default()
	server.GET("hello", func(context *gin.Context) {
		context.String(http.StatusOK, "你好")
	})
	server.Run(":8080")
}
