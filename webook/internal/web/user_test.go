package web

import (
	"fmt"
	"testing"
)

func TestTypeAssert(t *testing.T) {
	var id any
	if id == nil {
		fmt.Println("nil")
	}
	_, ok := id.(int64)
	if ok {
		fmt.Println("ok")
	} else {
		fmt.Println("!ok")
	}
}
