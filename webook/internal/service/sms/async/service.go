package async

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
	"webook/webook/internal/domain"
	repository2 "webook/webook/internal/repository/sms"
	"webook/webook/internal/service/sms"
	"webook/webook/pkg/ginx/ratelimit"
	"webook/webook/pkg/logger"
)

type SMSService struct {
	// 短信服务提供商
	sp []sms.Service
	// 存储请求
	repo    repository2.AsyncSMSRepository
	l       logger.Logger
	limiter ratelimit.Limiter
	// 限流对象，要用哪一家短信服务商，如腾讯云，则限制访问腾讯云服务的key访问次数
	key string

	// 服务商对应的数组下标
	idx int32
	// 连续超时的次数
	cnt int32

	// 阈值，连续超时超过这个数字，就要切换服务商
	thredshold int32
}

// StartAsyncCycle 异步发送消息
func (s *SMSService) StartAsyncCycle() {
	// 这个是我为了测试而引入的，防止你在运行测试的时候，会出现偶发性的失败
	//time.Sleep(time.Second * 3)

	for {
		s.AsyncSend()
	}
}

func (s *SMSService) AsyncSend() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// 抢占一个异步发送的消息，确保在非常多个实例,一个请求，只有一个实例能拿到
	as, err := s.repo.PreemptWaitingSMS(ctx)
	cancel()
	switch {
	case err == nil:
		// 执行发送
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		err = s.sendByFailOver(ctx, as.TplId, as.Args, as.Numbers...)
		if err != nil {
			s.l.Error("执行异步发送短信失败",
				logger.Error(err),
				logger.Int64("id", as.Id))
		}
		res := err == nil
		// 通知 repository 这一次的执行结果
		err = s.repo.ReportScheduleResult(ctx, as.Id, res)
		if err != nil {
			s.l.Error("执行异步发送短信成功，但是标记数据库失败",
				logger.Error(err),
				logger.Bool("res", res),
				logger.Int64("id", as.Id))
		}
	case errors.Is(err, repository2.ErrWaitingSMSNotFound):
		time.Sleep(time.Second)
	default:
		// 这里可能是数据库出了问题
		s.l.Error("抢占异步发送短信任务失败",
			logger.Error(err))
		time.Sleep(time.Second)
	}
}

func (s *SMSService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, s.key)
	if err != nil || limited {
		// 触发了限流需要转异步
		// 判断是否需要限流出现了问题，我也转为异步
		return errors.New("触发异步发送了")
	}
	// 把所有服务商都发一遍
	err = s.sendByFailOver(ctx, tplId, args, numbers...)
	if err == nil {
		return nil
	}
	// 如果都失败了，就异步
	return s.repo.Add(ctx, domain.SmsAsync{
		TplId:    tplId,
		Args:     args,
		Numbers:  numbers,
		RetryMax: 3,
	})
}

func (s *SMSService) sendByFailOver(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&s.idx)
	cnt := atomic.LoadInt32(&s.cnt)
	if cnt > s.thredshold {
		//超过了阈值，要切换，idx往后挪一位
		newIdx := (idx + 1) % int32(len(s.sp))
		if atomic.CompareAndSwapInt32(&s.idx, idx, newIdx) {
			// 成功后挪一位，就是切换了服务商，超时次数重置为0
			atomic.StoreInt32(&s.cnt, 0)
		}
		// else 就是出现并发，别人换成功了，我直接用别人的成果
		idx = atomic.LoadInt32(&s.idx)
	}
	svc := s.sp[idx]
	err := svc.Send(ctx, tplId, args, numbers...)
	switch err {
	case context.DeadlineExceeded:
		// 这个服务商超时了
		atomic.AddInt32(&s.cnt, 1)
		// 超时了，可以考虑换下一家
		return err
	case nil:
		// 成功了，该服务商的连续超时次数置为0
		atomic.StoreInt32(&s.cnt, 0)
		return nil
	default:
		// 不知道什么错误
		// 可以再换下一个服务商
		return err
	}
}
