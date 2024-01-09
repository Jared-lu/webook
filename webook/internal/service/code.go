package service

import (
	"context"
	"fmt"
	"math/rand"
	"webook/webook/internal/repository"
	"webook/webook/internal/service/sms"
)

// 我的短信验证码模板Id
const codeTplId string = "1921139"

var (
	ErrCodeSendTooMany   = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooMany = repository.ErrCodeVerifyTooMany
)

// CodeService 验证码服务
type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

// SmsCodeService 短信验证码服务
type SmsCodeService struct {
	repo repository.CodeRepository
	// 短信服务
	smsSvc sms.Service
}

func NewSmsCodeService(repo repository.CodeRepository, smsSvc sms.Service) CodeService {
	return &SmsCodeService{
		repo:   repo,
		smsSvc: smsSvc,
	}
}

func (s *SmsCodeService) Send(ctx context.Context, biz string, phone string) error {
	// 生成一个验证码（谁来生成）
	// 放入到Redis
	// 发出去
	code := s.generateCode()
	// 存放验证码
	err := s.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 只有个redis存储验证码通过后才能发送短信
	return s.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
}

func (s *SmsCodeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return s.repo.Verify(ctx, biz, phone, inputCode)
}

func (s *SmsCodeService) generateCode() string {
	// 6位数，[0,1000000)
	nums := rand.Intn(1000000)
	return fmt.Sprintf("%06d", nums)
}
