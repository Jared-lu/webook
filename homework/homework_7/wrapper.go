package homework_7

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webook/webook/pkg/logger"
)

var l logger.Logger

// WrapBody 统一处理错误
func WrapBody[T any](fn func(ctx *gin.Context, req T) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req T
		if err := ctx.Bind(&req); err != nil {
			return
		}

		// 执行业务逻辑
		res, err := fn(ctx, req)
		if err != nil {
			l.Error("业务逻辑执行出错",
				logger.String("path", ctx.Request.URL.Path),
				// 命中的路由
				logger.String("route", ctx.FullPath()),
				logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

// Result HTTP的返回内容
type Result struct {
	// 业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
