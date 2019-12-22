package transfer

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/stringify"
)

type OutputReference struct {
	storageKey   []byte
	transferHash Hash
	addressHash  address.Address
}

func NewOutputReference(transferHash Hash, addressHash address.Address) *OutputReference {
	return &OutputReference{
		storageKey:   append(transferHash[:], addressHash[:]...),
		transferHash: transferHash,
		addressHash:  addressHash,
	}
}

func (reference *OutputReference) GetStorageKey() []byte {
	return reference.storageKey
}

func (reference *OutputReference) String() string {
	return stringify.Struct("TransferOutputReference",
		stringify.StructField("transferHash", reference.transferHash),
		stringify.StructField("addressHash", reference.addressHash),
	)
}
