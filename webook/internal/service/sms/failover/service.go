package failover

import "webook/webook/internal/service/sms"

type SMSService struct {
	// svcs 多个短信服务商，这里可以是使用了限流的短信服务
	svcs []sms.Service
	idx  uint64
}
