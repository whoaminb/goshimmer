package ledgerstate

/*
import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/transfer"
	"github.com/iotaledger/goshimmer/packages/stringify"
)

type TransferOutputReference struct {
	storageKey   []byte
	transferHash transfer.Hash
	addressHash  address.Address
}

func NewTransferOutputReference(transferHash transfer.Hash, addressHash address.Address) *TransferOutputReference {
	return &TransferOutputReference{
		storageKey:   append(transferHash[:], addressHash[:]...),
		transferHash: transferHash,
		addressHash:  addressHash,
	}
}

func (transferOutputReference *TransferOutputReference) GetStorageKey() []byte {
	return transferOutputReference.storageKey
}

func (transferOutputReference *TransferOutputReference) String() string {
	return stringify.Struct("TransferOutputReference",
		stringify.StructField("transferHash", transferOutputReference.transferHash),
		stringify.StructField("addressHash", transferOutputReference.addressHash),
	)
}

*/
