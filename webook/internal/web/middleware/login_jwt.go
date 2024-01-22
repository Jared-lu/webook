package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	web2 "webook/webook/internal/web/jwt"
)

// LoginJWTMiddleWareBuilder 登录校验，使用JWT机制
type LoginJWTMiddleWareBuilder struct {
	// 不进行登录校验的路径
	paths []string
	web2.JWTHandler
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
		// 实现效果较差，可考虑改造成map结构存储
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		// 校验JWT Token，这里是短token校验，长token不会进来这里
		// 前端把token放到 Authorization 首部
		// 如果这里拿不到，后面的解析肯定失败
		claims := &web2.JWTUserClaims{} // 要用指针，因为要作为参数，让被掉函数修改再返回来
		// 我这里校验的是短token
		err := l.CheckToken(ctx, claims, web2.AtKey)
		//tokenStr := l.ExtractToken(ctx)
		//token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		//	// 我这里校验的是短token
		//	return web2.AtKey, nil
		//})
		if err != nil {
			// 没登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//if !token.Valid  {
		//	// 没登录
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		// 登录校验
		if claims.UserAgent != ctx.Request.UserAgent() || claims.Uid == 0 { // Uid是数据库自增主键，我们用了默认从1开始，不可能为0
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 方便业务要拿到这个数据
		ctx.Set("userId", claims.Uid)
		ctx.Set("claims", claims)

		// 判断有没有退出登录的ssid
		err = l.CheckSession(ctx, claims.Ssid)
		if err != nil {
			// 要么Redis有问题，要么已经退出登录了
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		////刷新jwt token
		//// 每一分钟刷一次
		//if claims.ExpiresAt.Sub(time.Now()) > time.Minute*29 {
		//	return
		//}
		//// 过期时间要重新设置一下
		//claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 30))
		//// 再重新生成token
		//tokenStr, err = token.SignedString([]byte("HiIilLa4O8Xy3Pm8C5mh5HymYaYt9eTj"))
		//if err != nil {
		//	log.Println("jwt 续约失败")
		//}
		//ctx.Header("x-jwt-token", tokenStr)
	}
}
