package homework_1

import (
	"errors"
)

// DeleteAt 删除切片特定下标的元素
func DeleteAt[T any](src []T, index int) ([]T, error) {
	length := len(src)
	if index >= length {
		return nil, errors.New("index out of range")
	}
	for i := index; i+1 < length; i++ {
		// 元素往前移
		src[i] = src[i+1]
	}
	return src, nil
}

// Contract 切片缩容
func Contract[T any](src []T) []T {
	l, c := len(src), cap(src)
	if c < 256 {
		// 容量小于256没必要缩
		return src
	}
	var s []T
	if c >= 1024 && c/l >= 2 {
		s = make([]T, 0, c/2)
		s = append(s, src...)
	}
	if c < 1024 && c/l >= 4 {
		s = make([]T, 0, c/2)
		s = append(s, src...)
	}
	return s
}
