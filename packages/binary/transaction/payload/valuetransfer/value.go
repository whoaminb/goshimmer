package valuetransfer

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
)

type ValueTransfer struct {
	inputs      []*transfer.OutputReference
	inputsMutex sync.RWMutex
}

var Type = payload.Type(1)

func New() *ValueTransfer {
	return &ValueTransfer{
		inputs: make([]*transfer.OutputReference, 0),
	}
}

func (valueTransfer *ValueTransfer) AddInput(transferHash transfer.Hash, address address.Address) *ValueTransfer {
	valueTransfer.inputsMutex.Lock()
	valueTransfer.inputs = append(valueTransfer.inputs, transfer.NewOutputReference(transferHash, address))
	valueTransfer.inputsMutex.Unlock()

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
