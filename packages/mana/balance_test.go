package mana

import (
	"fmt"
	"testing"
)

func TestBalance_AddTransfer(t *testing.T) {
	calculator := NewCalculator(500, 0.1)

	balance1 := NewBalance(calculator)
	balance1.AddTransfer(1000, 0, 500)
	balance1.AddTransfer(1000, 500, 1000)
	balance1.AddTransfer(1000, 1000, 1700)
	fmt.Println(balance1.GetValue())

	balance2 := NewBalance(calculator)
	balance2.AddTransfer(1000, 0, 1700)
	fmt.Println(balance2.GetValue())
}
