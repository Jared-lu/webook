package tencent

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

// SmsService 短信服务
type SmsService struct {
	// 应用ID
	appId *string
	// 签名，说明是谁发送的
	signName *string
	// 短信服务商
	sp *sms.Client
}

func NewSMSService(sp *sms.Client, appId string, signName string) *SmsService {
	return &SmsService{
		sp: sp,
		// 转换为对应类型的指针类型
		// 因为腾讯的API要求传的就是指针
		appId:    &appId,
		signName: &signName,
	}

}

// Send 腾讯云发送短信
// templateId 模板Id
func (s *SmsService) Send(ctx context.Context, templateId string, args []string, numbers ...string) error {
	req := sms.NewSendSmsRequest()
	// 设置要求传入的参数
	req.SmsSdkAppId = s.appId
	req.SignName = s.signName
	// 短信模板Id
	req.TemplateId = &templateId
	// 这下面也要转换成对应的指针
	// 这个短信API设计的真够恶心的
	req.PhoneNumberSet = s.toStringPtrSlice(numbers)
	// 传给模板中参数的值，数量少要对应
	req.TemplateParamSet = s.toStringPtrSlice(args)
	// 发送短信
	resp, err := s.sp.SendSms(req)
	if err != nil {
		return err
	}
	// 一条短信一个number，有些手机发成功了，有些可能没有成功，需要逐个解析是否全部成功
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送短信失败 %s %s ", *status.Code, *status.Message)
		}
	}
	return nil

}

func (s *SmsService) toStringPtrSlice(src []string) []*string {
	return slice.Map[string, *string](src, func(idx int, src string) *string {
		return &src
	})
}
