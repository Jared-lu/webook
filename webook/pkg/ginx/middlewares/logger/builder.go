package logger

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

// LoggerBuilder 打印请求和响应的Middleware
type LoggerBuilder struct {
	// 是否打印Request Body
	allowReqBody bool
	// 是否打印Response Body
	allowRespBody bool
	// 记录日志的具体方法
	loggerFunc func(ctx context.Context, al *AccessLog)
}

func NewLoggerBuilder(loggerFunc func(ctx context.Context, al *AccessLog)) *LoggerBuilder {
	return &LoggerBuilder{loggerFunc: loggerFunc}
}

func (l *LoggerBuilder) AllowReqBody() *LoggerBuilder {
	l.allowReqBody = true
	return l
}

func (l *LoggerBuilder) AllowRespBody() *LoggerBuilder {
	l.allowRespBody = true
	return l
}

func (l *LoggerBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		// 这里就是要执行打印了
		// 至于怎么打，让调用者自己决定
		url := ctx.Request.URL.String()
		// 只打印前1024个字节
		if len(url) > 1024 {
			url = url[:1024]
		}
		al := &AccessLog{
			Method: ctx.Request.Method,
			Url:    url,
		}
		// 记录请求
		if l.allowReqBody && ctx.Request.Body != nil {
			body, _ := ctx.GetRawData()
			// 要把body放回去
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			if len(body) > 1024 {
				body = body[:1024]
			}
			// 这是一个复制操作，很消耗性能和内存
			al.ReqBody = string(body)
		}
		// 记录响应
		if l.allowRespBody {
			// 这里是为了让GIN构造http响应时，把我需要的内容写入到 AccessLog，
			// 然后再返回响应前还要打印出来，因此一定是要用 AccessLog 指针
			ctx.Writer = responseWriter{
				al: al,
				// 这里是初始化没有重写的方法
				ResponseWriter: ctx.Writer,
			}
		}

		// 执行业务逻辑
		ctx.Next()

		defer func() {
			al.Duration = time.Since(start).String()
			// 一定是执行完业务逻辑才能打印，不然是没有响应内容的
			l.loggerFunc(ctx, al)
		}()

	}
}

func (w responseWriter) WriteHeader(statusCode int) {
	w.al.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
func (w responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w responseWriter) WriteString(data string) (int, error) {
	w.al.RespBody = data
	return w.ResponseWriter.WriteString(data)
}

// responseWriter 装饰 gin.ResponseWriter，为了能够拿到http 响应
type responseWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

// AccessLog 表示这一次请求要打印到日志中的内容
type AccessLog struct {
	Method   string
	Url      string
	ReqBody  string
	RespBody string
	// 响应码
	Status int
	// 这一个请求的处理时间
	Duration string
}
