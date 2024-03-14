package domain

type SmsAsync struct {
	Id      int64
	TplId   string
	Args    []string
	Numbers []string
	// 允许重试的最大次数
	RetryMax int
}
