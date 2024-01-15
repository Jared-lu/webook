package ioc

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentSms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
	"webook/webook/internal/service/sms"
	"webook/webook/internal/service/sms/memory"
	"webook/webook/internal/service/sms/tencent"
)

func InitSMSService() sms.Service {
	return initMemorySMSService()
}

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

func initMemorySMSService() sms.Service {
	return memory.NewSmsService()
}
