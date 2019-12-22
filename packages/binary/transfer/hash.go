package transfer

import (
	"unicode/utf8"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"

	"github.com/iotaledger/goshimmer/packages/stringify"
)

type Hash [HashLength]byte

func NewHash(transferHash string) (result Hash) {
	copy(result[:], transferHash)

	return
}

func (transferHash Hash) ToRealityId() (realityId reality.Id) {
	copy(realityId[:], transferHash[:])

	return
}

func (transferHash *Hash) UnmarshalBinary(data []byte) error {
	copy(transferHash[:], data[:HashLength])

	return nil
}

func (transferHash Hash) String() string {
	if utf8.Valid(transferHash[:]) {
		return string(transferHash[:])
	} else {
		return stringify.SliceOfBytes(transferHash[:])
	}
}

const HashLength = 32
