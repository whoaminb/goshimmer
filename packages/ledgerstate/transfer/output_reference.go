package transfer

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/stringify"
)

type OutputReference struct {
	storageKey   []byte
	transferHash Id
	addressHash  address.Address
}

func NewOutputReference(transferHash Id, addressHash address.Address) *OutputReference {
	return &OutputReference{
		storageKey:   append(transferHash[:], addressHash[:]...),
		transferHash: transferHash,
		addressHash:  addressHash,
	}
}

func (reference *OutputReference) GetTransferHash() Id {
	return reference.transferHash
}

func (reference *OutputReference) GetAddress() address.Address {
	return reference.addressHash
}

func (reference *OutputReference) MarshalBinary() (result []byte, err error) {
	result = make([]byte, IdLength+address.Length)
	offset := 0

	copy(result[offset:], reference.transferHash[:])
	offset += IdLength

	copy(result[offset:], reference.addressHash[:])

	return
}

func (reference *OutputReference) UnmarshalBinary(bytes []byte) (err error) {
	offset := 0

	copy(reference.transferHash[:], bytes[offset:])
	offset += IdLength

	copy(reference.addressHash[:], bytes[offset:])

	return
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
