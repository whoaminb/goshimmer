package ledgerstate

import (
	"unicode/utf8"

	"github.com/iotaledger/goshimmer/packages/stringify"
)

type RealityId [realityIdLength]byte

func NewRealityId(realityId string) (result RealityId) {
	copy(result[:], realityId)

	return
}

func (realityId *RealityId) UnmarshalBinary(data []byte) error {
	copy(realityId[:], data[:realityIdLength])

	return nil
}

func (realityId RealityId) String() string {
	if utf8.Valid(realityId[:]) {
		return string(realityId[:])
	} else {
		return stringify.SliceOfBytes(realityId[:])
	}
}

var emptyRealityId = RealityId{}

const realityIdLength = 32
