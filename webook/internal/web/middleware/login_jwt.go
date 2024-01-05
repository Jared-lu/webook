package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

type LoginJWTMiddleWareBuilder struct {
	// 不进行登录校验的路径
	paths []string
}

func NewLoginJWTMiddleWareBuilder() *LoginJWTMiddleWareBuilder {
	return &LoginJWTMiddleWareBuilder{}
}

func (l *LoginJWTMiddleWareBuilder) IgnorePaths(path string) *LoginJWTMiddleWareBuilder {
	l.paths = append(l.paths, path)
	return l
}

// Build 也可以叫CheckLogin
func (l *LoginJWTMiddleWareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 实现效果较差，
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// 校验JWT Token
		// 前端把token放到 Authorization 首部
		tokenHeader := ctx.GetHeader("Authorization")
		if tokenHeader == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		sets := strings.Split(tokenHeader, " ")
		if len(sets) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := sets[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte("HiIilLa4O8Xy3Pm8C5mh5HymYaYt9eTj"), nil
		})
		if err != nil {
			// 没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if !token.Valid {
			// 没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
