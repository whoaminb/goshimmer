package transferoutput

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
	"github.com/iotaledger/goshimmer/packages/stringify"
)

type Reference struct {
	storageKey   []byte
	transferHash transfer.Hash
	addressHash  address.Address
}

func NewTransferOutputReference(transferHash transfer.Hash, addressHash address.Address) *Reference {
	return &Reference{
		storageKey:   append(transferHash[:], addressHash[:]...),
		transferHash: transferHash,
		addressHash:  addressHash,
	}
}

func (reference *Reference) GetStorageKey() []byte {
	return reference.storageKey
}

func (reference *Reference) String() string {
	return stringify.Struct("TransferOutputReference",
		stringify.StructField("transferHash", reference.transferHash),
		stringify.StructField("addressHash", reference.addressHash),
	)
}
