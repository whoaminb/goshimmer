package coloredbalance

import (
	"strconv"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
)

type ColoredBalance struct {
	color   hash.Color
	balance uint64
}

func New(color hash.Color, balance uint64) *ColoredBalance {
	return &ColoredBalance{
		color:   color,
		balance: balance,
	}
}

func (balance *ColoredBalance) GetColor() hash.Color {
	return balance.color
}

func (balance *ColoredBalance) GetValue() uint64 {
	return balance.balance
}

func (coloredBalance *ColoredBalance) String() string {
	return "ColoredBalance(\"" + coloredBalance.color + "\", " + strconv.FormatUint(coloredBalance.balance, 10) + ")"
}
