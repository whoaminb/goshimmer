package ledgerstate

import (
	"strconv"
)

type ColoredBalance struct {
	color   ColorHash
	balance uint64
}

func NewColoredBalance(color ColorHash, balance uint64) *ColoredBalance {
	return &ColoredBalance{
		color:   color,
		balance: balance,
	}
}

func (coloredBalance *ColoredBalance) String() string {
	return "ColoredBalance(\"" + coloredBalance.color + "\", " + strconv.FormatUint(coloredBalance.balance, 10) + ")"
}
