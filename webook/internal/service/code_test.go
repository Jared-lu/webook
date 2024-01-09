package service

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestSmsCodeService_generateCode(t *testing.T) {
	nums := rand.Intn(1000000)
	fmt.Printf("%06d", nums)
}
