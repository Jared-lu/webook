package logger

// 这里提供一些便利性的转换方法，不需要用户手动构造 Field

func String(key string, value string) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}

func Int64(key string, value int64) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}

func Error(err error) Field {
	return Field{
		Key:   "error",
		Value: err,
	}
}

func Bool(res string, val bool) Field {
	return Field{
		Key:   res,
		Value: val,
	}
}
