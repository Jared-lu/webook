package web

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type jwtHandler struct {
}

func (u jwtHandler) setJWTToken(ctx *gin.Context, id int64) error {
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
	tokenStr, err := token.SignedString([]byte("HiIilLa4O8Xy3Pm8C5mh5HymYaYt9eTj"))
	if err != nil {
		return err
	}
	// 将jwt token返回给前端，通过首部的方式
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

// JWTUserClaims JWT用户数据
type JWTUserClaims struct {
	jwt.RegisteredClaims // 实现Claims接口
	// 放入到token里的数据
	Uid       int64
	UserAgent string
}
