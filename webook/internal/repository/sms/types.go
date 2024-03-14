package repository

import (
	"context"
	"webook/webook/internal/domain"
)

type AsyncSMSRepository interface {
	Add(ctx context.Context, smsAsync domain.SmsAsync) error
	PreemptWaitingSMS(ctx context.Context) (domain.SmsAsync, error)
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}
