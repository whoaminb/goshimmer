package mana

import (
	"fmt"
	"testing"
)

func TestManaOfTransfer(t *testing.T) {
	fmt.Println(ManaOfTransfer(50, 10, 110))
	fmt.Println(ManaOfTransfer(50, 0, 190))
	fmt.Println(ManaOfTransfer(50, 0, 200))
	fmt.Println(ManaOfTransfer(50, 0, 290))
	fmt.Println(ManaOfTransfer(50, 0, 300))
	fmt.Println(ManaOfTransfer(50, 0, 390))
	fmt.Println(ManaOfTransfer(50, 0, 400))
}
