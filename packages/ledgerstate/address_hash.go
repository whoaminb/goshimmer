package ledgerstate

import (
	"unicode/utf8"

	"github.com/iotaledger/goshimmer/packages/stringify"
)

type AddressHash [addressHashLength]byte

func NewAddressHash(addressHash string) (result AddressHash) {
	copy(result[:], addressHash)

	return
}

func (addressHash *AddressHash) UnmarshalBinary(data []byte) error {
	copy(addressHash[:], data[:addressHashLength])

	return nil
}

func (addressHash AddressHash) String() string {
	if utf8.Valid(addressHash[:]) {
		return string(addressHash[:])
	} else {
		return stringify.SliceOfBytes(addressHash[:])
	}
}

const addressHashLength = 32
