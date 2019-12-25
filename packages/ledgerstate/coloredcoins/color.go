package coloredcoins

import (
	"unicode/utf8"

	"github.com/iotaledger/goshimmer/packages/stringify"
)

type Color [ColorLength]byte

func NewColor(color string) (result Color) {
	copy(result[:], color)

	return
}

func (color *Color) MarshalBinary() (result []byte, err error) {
	result = make([]byte, ColorLength)
	copy(result, color[:])

	return
}

func (color *Color) UnmarshalBinary(data []byte) error {
	copy(color[:], data[:ColorLength])

	return nil
}

func (color Color) String() string {
	if utf8.Valid(color[:]) {
		return string(color[:])
	} else {
		return stringify.SliceOfBytes(color[:])
	}
}

const ColorLength = 32
