package memory

import (
	"context"
	"fmt"
	"webook/webook/internal/service/sms"
)

type SmsService struct {
}

func NewSmsService() sms.Service {
	return &SmsService{}
}

func (s SmsService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	fmt.Println(args)
	return nil
}
