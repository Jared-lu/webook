package sms

import "context"

type Service interface {
	// Send
	// tplId 模板Id
	// args 具体的参数
	// numbers 发送的号码
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}
