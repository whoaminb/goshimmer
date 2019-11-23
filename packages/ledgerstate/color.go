package ledgerstate

import (
	"unicode/utf8"

	"github.com/iotaledger/goshimmer/packages/stringify"
)

type Color [colorLength]byte

func NewColor(color string) (result Color) {
	copy(result[:], color)

	return
}

func (color *Color) UnmarshalBinary(data []byte) error {
	copy(color[:], data[:colorLength])

	return nil
}

func (color Color) String() string {
	if utf8.Valid(color[:]) {
		return string(color[:])
	} else {
		return stringify.SliceOfBytes(color[:])
	}
}

const colorLength = 32
