package repository

import (
	"context"
	"github.com/ecodeclub/ekit/sqlx"
	"webook/webook/internal/domain"
	dao "webook/webook/internal/repository/dao/sms"
)

var ErrWaitingSMSNotFound = dao.ErrWaitingSMSNotFound

type SmsAsyncRepository struct {
	dao dao.AsyncSmsDAO
}

func NewSmsAsyncRepository(dao dao.AsyncSmsDAO) AsyncSMSRepository {
	return &SmsAsyncRepository{dao: dao}
}

func (a *SmsAsyncRepository) Add(ctx context.Context, s domain.SmsAsync) error {
	return a.dao.Insert(ctx, dao.AsyncSms{
		Config: sqlx.JsonColumn[dao.SmsConfig]{
			Val: dao.SmsConfig{
				TplId:   s.TplId,
				Args:    s.Args,
				Numbers: s.Numbers,
			},
			Valid: true,
		},
		RetryMax: s.RetryMax,
	})
}

// PreemptWaitingSMS 获取一个等待异步发送的短信请求
func (a *SmsAsyncRepository) PreemptWaitingSMS(ctx context.Context) (domain.SmsAsync, error) {
	as, err := a.dao.GetWaitingSMS(ctx)
	if err != nil {
		return domain.SmsAsync{}, err
	}
	return domain.SmsAsync{
		Id:       as.Id,
		TplId:    as.Config.Val.TplId,
		Numbers:  as.Config.Val.Numbers,
		Args:     as.Config.Val.Args,
		RetryMax: as.RetryMax,
	}, nil
}

// ReportScheduleResult 设置调度结果
func (a *SmsAsyncRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return a.dao.MarkSuccess(ctx, id)
	}
	return a.dao.MarkFailed(ctx, id)
}
