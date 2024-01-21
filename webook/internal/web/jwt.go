package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
)

type jwtHandler struct {
	// 短token的key
	acKey []byte
	// 长token的key
	rtkey []byte
}

func newJwtHandler() jwtHandler {
	return jwtHandler{
		acKey: []byte("HiIilLa4O8Xy3Pm8C5mh5HymYaYt9eTj"),
		rtkey: []byte("HiIilLa4O8Xy3Pm8C5mh5HymYaYt9eTj"),
	}
}

func (h jwtHandler) setJWTToken(ctx *gin.Context, id int64) error {
	// 生成JWT token`
	// JWT 带上个人数据作为一个身份识别
	claims := JWTUserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			// 设置jwt token的过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       id,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString(h.acKey)
	if err != nil {
		return err
	}
	// 将jwt token返回给前端，通过首部的方式
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

// setRefreshToken 刷新长token
func (h jwtHandler) setRefreshToken(ctx *gin.Context, uid int64) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			//有效期7天
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 60 * 24 * 7)),
		},
		Uid: uid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	// 设置token值
	tokenStr, err := token.SignedString(h.rtkey)
	if err != nil {
		return err
	}
	// 将token放到响应头部
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

// ExtractToken 获取token
func ExtractToken(ctx *gin.Context) string {
	// 长短token都放在 Authorization 首部，只有更新短token时这里面存放的才是长token，其余都是放token
	tokenHeader := ctx.GetHeader("Authorization")
	sets := strings.Split(tokenHeader, " ")
	if len(sets) != 2 {
		return ""
	}
	return sets[1]
}

type RefreshClaims struct {
	jwt.RegisteredClaims // 实现Claims接口
	// token中要带上的数据
	Uid int64
}

// JWTUserClaims JWT用户数据
type JWTUserClaims struct {
	jwt.RegisteredClaims // 实现Claims接口
	// 放入到token里的数据
	Uid       int64
	UserAgent string
}
