package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
	"time"
	"webook/webook/internal/service/sms"
	"webook/webook/internal/service/sms/memory"
	"webook/webook/internal/service/sms/ratelimit"
	"webook/webook/internal/service/sms/tencent"
	ratelimit2 "webook/webook/pkg/ginx/ratelimit"
)

//func InitSMSService() sms.Service {
//	return initMemorySMSService()
//}

func InitSMSService(cmd redis.Cmdable) sms.Service {
	return initRateLimitSmsService(cmd)
}

// 腾讯云短信服务
func initTencentSMSService() sms.Service {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("找不到环境变量 SMS_SECRET_ID")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")

	c, err := tencentSms.NewClient(common.NewCredential(secretId, secretKey),
		"ap-guangzhou", profile.NewClientProfile())
	if err != nil {
		panic("找不到环境变量 SMS_SECRET_KEY")
	}
	return tencent.NewSMSService(c, "1400853424", "猜猜我是谁")
}

// 内存实现
func initMemorySMSService() sms.Service {
	return memory.NewSmsService()
}

func initSMSLimiter(cmd redis.Cmdable) ratelimit2.Limiter {
	// 每秒限流3000个，这是腾讯的限流规则
	// 如果是其它云服务商，限流规则不同就要使用另外的限流器
	return ratelimit2.NewRedisSlidingWindowLimiter(cmd, time.Second, 3000)
}

func initRateLimitSmsService(cmd redis.Cmdable) sms.Service {
	// 每一家短信服务都有各自的限流规则
	return ratelimit.NewService(initTencentSMSService(), initSMSLimiter(cmd), "TencentSMS")
}
