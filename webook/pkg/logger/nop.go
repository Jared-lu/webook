package logger

// NopLogger 这个是什么都不做的实现，一般用于测试方便
type NopLogger struct {
}

func NewNoOpLogger() Logger {
	return &NopLogger{}
}

func (n *NopLogger) Debug(msg string, args ...Field) {

}

func (n *NopLogger) Info(msg string, args ...Field) {

}

func (n *NopLogger) Warn(msg string, args ...Field) {

}

func (n *NopLogger) Error(msg string, args ...Field) {

}
