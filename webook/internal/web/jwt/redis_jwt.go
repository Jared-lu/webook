package web

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var (
	// AtKey access_token key
	AtKey = []byte("HiIilLa4O8Xy3Pm8C5mh5HymYaYt9eTj")
	// RtKey refresh_token key
	RtKey = []byte("HiIilLa4O8Xy3Pm8C5mh5HymYaYt9eTj")
)

var ErrTokenInvalid = errors.New("无效的token")

type RedisJWTHandler struct {
	cmd redis.Cmdable
}

func NewRedisJWTHandler(cmd redis.Cmdable) JWTHandler {
	return &RedisJWTHandler{cmd: cmd}
}

func (r *RedisJWTHandler) CheckToken(ctx *gin.Context, claims jwt.Claims, key []byte) error {
	tokenStr := r.ExtractToken(ctx)
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return ErrTokenInvalid
	}
	return nil
}

func (r *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	// 约定长短token都放在 Authorization 首部，只有更新短token时这里面存放的才是长token，其余都是放token
	tokenHeader := ctx.GetHeader("Authorization")
	sets := strings.Split(tokenHeader, " ")
	if len(sets) != 2 {
		return ""
	}
	return sets[1]
}

func (r *RedisJWTHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := r.SetJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	return r.SetRefreshToken(ctx, uid, ssid)
}

func (r *RedisJWTHandler) SetJWTToken(ctx *gin.Context, uid int64, ssid string) error {
	// 生成 JWT token`
	// JWT 带上个人数据作为一个身份识别
	claims := JWTUserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			// 设置jwt token的过期时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
		Uid:       uid,
		Ssid:      ssid,
		UserAgent: ctx.Request.UserAgent(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	// 生成access_token
	tokenStr, err := token.SignedString(AtKey)
	if err != nil {
		return err
	}
	// 将jwt token返回给前端，通过首部的方式
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

func (r *RedisJWTHandler) SetRefreshToken(ctx *gin.Context, uid int64, ssid string) error {
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			//有效期7天
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 60 * 24 * 7)),
		},
		Uid:  uid,
		Ssid: ssid,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	// 设置refresh_token
	tokenStr, err := token.SignedString(RtKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (r *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	claims := ctx.MustGet("claims").(*JWTUserClaims)
	// 退出登录的关键就是将ssid标记为不可用，如果Redis中存在这个ssid，说明用户已经退出登录了
	return r.cmd.Set(ctx, fmt.Sprintf("user:ssid:%s", claims.Ssid), "", time.Hour*24*7).Err()
}

func (r *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	val, err := r.cmd.Exists(ctx, fmt.Sprintf("users:ssid:%s", ssid)).Result()
	switch {
	case errors.Is(err, redis.Nil):
		return nil
	case err == nil:
		if val == 0 {
			return nil
		}
		return errors.New("已经无效了")
	default:
		return err
	}
}

type RefreshClaims struct {
	jwt.RegisteredClaims // 实现Claims接口
	// token中要带上的数据
	Uid  int64
	Ssid string
}

// JWTUserClaims JWT用户数据
type JWTUserClaims struct {
	jwt.RegisteredClaims // 实现Claims接口
	// 放入到token里的数据
	Uid       int64
	Ssid      string
	UserAgent string
}
