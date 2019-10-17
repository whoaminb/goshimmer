package ledgerstate

import (
	"strconv"
)

type Color string

type ColoredBalance struct {
	color   Color
	balance uint64
}

func NewColoredBalance(color Color, balance uint64) *ColoredBalance {
	return &ColoredBalance{
		color:   color,
		balance: balance,
	}
}

func (balance *ColoredBalance) GetColor() Color {
	return balance.color
}

func (balance *ColoredBalance) GetValue() uint64 {
	return balance.balance
}

func (coloredBalance *ColoredBalance) String() string {
	return "ColoredBalance(\"" + string(coloredBalance.color) + "\", " + strconv.FormatUint(coloredBalance.balance, 10) + ")"
}
