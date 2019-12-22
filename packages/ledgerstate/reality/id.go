package reality

import (
	"unicode/utf8"

	"github.com/iotaledger/goshimmer/packages/stringify"
)

type Id [IdLength]byte

func NewId(realityId string) (result Id) {
	copy(result[:], realityId)

	return
}

func (id *Id) UnmarshalBinary(data []byte) error {
	copy(id[:], data[:IdLength])

	return nil
}

func (id Id) String() string {
	if utf8.Valid(id[:]) {
		return string(id[:])
	} else {
		return stringify.SliceOfBytes(id[:])
	}
}

var EmptyId = Id{}

const IdLength = 32
