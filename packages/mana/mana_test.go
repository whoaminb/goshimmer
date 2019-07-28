package mana

import (
	"fmt"
	"testing"
)

func TestManaOfTransfer(t *testing.T) {
	fmt.Println(ManaOfTransfer(50, 99))
	fmt.Println(ManaOfTransfer(50, 100))
	fmt.Println(ManaOfTransfer(50, 200))
	fmt.Println(ManaOfTransfer(50, 290))
	fmt.Println(ManaOfTransfer(50, 300))
}
