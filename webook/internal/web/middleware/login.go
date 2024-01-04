package middleware

import (
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddleWareBuilder struct {
	// 不进行登录校验的路径
	paths []string
}

func NewLoginMiddleWareBuilder() *LoginMiddleWareBuilder {
	return &LoginMiddleWareBuilder{}
}

func (l *LoginMiddleWareBuilder) IgnorePaths(path string) *LoginMiddleWareBuilder {
	l.paths = append(l.paths, path)
	return l
}

// Build 也可以叫CheckLogin
func (l *LoginMiddleWareBuilder) Build() gin.HandlerFunc {
	// 用Go的方式编码解码，缺少会panic
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		// 实现效果较差，
		for _, path := range l.paths {
			if ctx.Request.URL.Path == path {
				return
			}
		}

		sess := sessions.Default(ctx)
		// 前面已经设置了session，因此sess不可能为nil
		//if sess == nil {
		//	// 没有登录
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		// 获取session的值
		id := sess.Get("userId")
		if id == nil {
			// 没有登录
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// 下面要刷新Session id
		// 先拿到上次的更新时间
		updateTime := sess.Get("update_time")
		// Session的数据值也需要更新，因为数据和sess_id 放到一块，不然后面的请求就拿不到session里面的数据
		sess.Set("userId", id)
		// 每刷新一次都要重置一下cookie的过期时间
		sess.Options(sessions.Options{
			MaxAge: 10 * 60, // 设置过期时间 30 * 60s（演示效果10*60s)
		})
		now := time.Now().UnixMilli()
		if updateTime == nil {
			// 说明还没有刷新过，即刚登录后的第一次请求，还没刷新
			// 刷新一下session id
			sess.Set("update_time", now)
			sess.Save()
			return
		}
		// 防止黑客
		updateTimeVal, ok := updateTime.(int64)
		// 这里可不校验
		if !ok {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// 刷新频率：每分钟刷一次
		if now-updateTimeVal > 10*1000 { // 60000毫秒=60秒（演示效果用10秒）
			// 刷新一下
			sess.Set("update_time", now)
			sess.Save()
			return
		}
	}
}
