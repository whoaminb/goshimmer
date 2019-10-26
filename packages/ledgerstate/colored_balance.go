package ledgerstate

import (
	"encoding/binary"
	"strconv"
)

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

func (balance *ColoredBalance) UnmarshalBinary(data []byte) error {
	balance.color = Color{}
	if err := balance.color.UnmarshalBinary(data); err != nil {
		return err
	}

	balance.balance = binary.LittleEndian.Uint64(data[colorLength:])

	return nil
}

func (coloredBalance *ColoredBalance) String() string {
	return "ColoredBalance(\"" + coloredBalance.color.String() + "\", " + strconv.FormatUint(coloredBalance.balance, 10) + ")"
}

const (
	coloredBalanceLength = colorLength + 8 // color + 64 Bit / 8 Byte value
)
