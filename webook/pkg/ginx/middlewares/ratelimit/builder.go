package ratelimit

import (
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"webook/webook/pkg/ginx/ratelimit"
)

type Builder struct {
	prefix string
	//cmd      redis.Cmdable
	//interval time.Duration
	//// 阈值
	//rate    int
	limiter ratelimit.Limiter
}

func NewBuilder(limiter ratelimit.Limiter) *Builder {
	return &Builder{prefix: "ip-limiter", limiter: limiter}
}

// Prefix 如果不用ip作为限流对象，则自己传进来
func (b *Builder) Prefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limit(ctx)
		if err != nil {
			log.Println(err)
			// 这一步很有意思，就是如果这边出错了
			// 要怎么办？
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limited {
			log.Println(err)
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}

func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	key := fmt.Sprintf("%s:%s", b.prefix, ctx.ClientIP())
	return b.limiter.Limit(ctx.Request.Context(), key)
}
