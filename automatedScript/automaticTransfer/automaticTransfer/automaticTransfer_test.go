package automaticTransfer

import (
	"testing"
	"fmt"
	"math/big"
)

func TestGetBalance(t *testing.T) {
	result := GetBalance("0xebb97ad3ca6b4f609da161c0b2b0eaa4ad58f3e8")
	fmt.Println(result)
	result.Sub(result,big.NewInt(5000000000000000000))
	if result.Cmp(big.NewInt(0)) < 0 {
		return
	}
	fmt.Println(result)
}

func TestAutomaticTransfer(t *testing.T) {
	AutomaticTransfer("0xebb97ad3ca6b4f609da161c0b2b0eaa4ad58f3e8")
}