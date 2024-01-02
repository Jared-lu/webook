package web

// Result HTTP的返回内容
type Result struct {
	// 业务错误码
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
