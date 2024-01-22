package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTHandler JWT有关的
type JWTHandler interface {
	// ExtractToken 获取token
	ExtractToken(ctx *gin.Context) string
	// SetLoginToken 设置登录态
	SetLoginToken(ctx *gin.Context, uid int64) error
	// SetJWTToken 长短token机制下，设置短token，其余时候用作生成jwt token
	// ssid 为当前session
	SetJWTToken(ctx *gin.Context, uid int64, ssid string) error
	// SetRefreshToken 长短token机制下，设置长Token
	SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error
	// ClearToken 清理jwt token
	ClearToken(ctx *gin.Context) error
	// CheckSession 检测Session是否存在，用于退出登录等
	CheckSession(ctx *gin.Context, ssid string) error
	// CheckToken 校验token是否有效
	CheckToken(ctx *gin.Context, claims jwt.Claims, key []byte) error
}
