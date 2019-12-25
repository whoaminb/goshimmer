package coloredcoins

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

func (coloredBalance *ColoredBalance) GetColor() Color {
	return coloredBalance.color
}

func (coloredBalance *ColoredBalance) GetBalance() uint64 {
	return coloredBalance.balance
}

func (coloredBalance *ColoredBalance) MarshalBinary() (result []byte, err error) {
	result = make([]byte, ColorLength+8)

	if marshaledColor, marshalErr := coloredBalance.color.MarshalBinary(); marshalErr != nil {
		err = marshalErr

		return
	} else {
		copy(result, marshaledColor)
	}

	binary.LittleEndian.PutUint64(result[ColorLength:], coloredBalance.balance)

	return
}

func (coloredBalance *ColoredBalance) UnmarshalBinary(data []byte) error {
	coloredBalance.color = Color{}
	if err := coloredBalance.color.UnmarshalBinary(data); err != nil {
		return err
	}

	coloredBalance.balance = binary.LittleEndian.Uint64(data[ColorLength:])

	return nil
}

func (coloredBalance *ColoredBalance) String() string {
	return "ColoredBalance(\"" + coloredBalance.color.String() + "\", " + strconv.FormatUint(coloredBalance.balance, 10) + ")"
}

const (
	BalanceLength = ColorLength + 8 // color + 64 Bit / 8 Byte value
)
