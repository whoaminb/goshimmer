package transfer

import (
	"unicode/utf8"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"
	"github.com/iotaledger/goshimmer/packages/stringify"
)

type Id [IdLength]byte

func NewId(id []byte) (result Id) {
	copy(result[:], id)

	return
}

func (transferId Id) ToRealityId() (realityId reality.Id) {
	copy(realityId[:], transferId[:])

	return
}

func (transferId *Id) MarshalBinary() (result []byte, err error) {
	result = make([]byte, IdLength)
	copy(result, transferId[:])

	return
}

func (transferId *Id) UnmarshalBinary(data []byte) error {
	copy(transferId[:], data[:IdLength])

	return nil
}

func (transferId Id) String() string {
	if utf8.Valid(transferId[:]) {
		return string(transferId[:])
	} else {
		return stringify.SliceOfBytes(transferId[:])
	}
}

var EmptyId = Id{}

const IdLength = 32
