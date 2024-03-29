package sms

import "context"

// Service 短信服务
type Service interface {
	// Send
	// tplId 模板Id
	// args 模板的占位符参数
	// numbers 发送的号码
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}
