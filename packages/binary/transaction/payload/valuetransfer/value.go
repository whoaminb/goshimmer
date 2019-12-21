package valuetransfer

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload"
)

type ValueTransfer struct{}

var Type = payload.Type(1)

func New() *ValueTransfer {
	return &ValueTransfer{}
}

func (valueTransfer *ValueTransfer) AddInput(transaction transaction.Id, address address.Address) *ValueTransfer {
	return valueTransfer
}

func (valueTransfer *ValueTransfer) GetType() payload.Type {
	return Type
}

func (valueTransfer *ValueTransfer) MarshalBinary() (bytes []byte, err error) {
	return
}

func (valueTransfer *ValueTransfer) UnmarshalBinary(bytes []byte) (err error) {
	return
}
