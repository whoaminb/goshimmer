package transfer

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/coloredcoins"
)

type Transfer struct {
	id      Id
	inputs  []*OutputReference
	outputs map[address.Address][]*coloredcoins.ColoredBalance
}

func NewTransfer(id Id) *Transfer {
	return &Transfer{
		id:      id,
		inputs:  make([]*OutputReference, 0),
		outputs: make(map[address.Address][]*coloredcoins.ColoredBalance),
	}
}

func (transfer *Transfer) GetId() Id {
	return transfer.id
}

func (transfer *Transfer) GetInputs() []*OutputReference {
	return transfer.inputs
}

func (transfer *Transfer) AddInput(input *OutputReference) *Transfer {
	transfer.inputs = append(transfer.inputs, input)

	return transfer
}

func (transfer *Transfer) AddOutput(address address.Address, balance *coloredcoins.ColoredBalance) *Transfer {
	if _, addressExists := transfer.outputs[address]; !addressExists {
		transfer.outputs[address] = make([]*coloredcoins.ColoredBalance, 0)
	}

	transfer.outputs[address] = append(transfer.outputs[address], balance)

	return transfer
}

func (transfer *Transfer) GetOutputs() map[address.Address][]*coloredcoins.ColoredBalance {
	return transfer.outputs
}

func (transfer *Transfer) MarshalBinary() (data []byte, err error) {
	return
}

func (transfer *Transfer) UnmarshalBinary(data []byte) (err error) {
	return
}
