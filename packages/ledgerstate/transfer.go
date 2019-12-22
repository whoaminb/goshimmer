package ledgerstate

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/transfer"
	"github.com/iotaledger/goshimmer/packages/binary/transferoutput"
)

type Transfer struct {
	hash    transfer.Hash
	inputs  []*transferoutput.Reference
	outputs map[address.Address][]*ColoredBalance
}

func NewTransfer(transferHash transfer.Hash) *Transfer {
	return &Transfer{
		hash:    transferHash,
		inputs:  make([]*transferoutput.Reference, 0),
		outputs: make(map[address.Address][]*ColoredBalance),
	}
}

func (transfer *Transfer) GetHash() transfer.Hash {
	return transfer.hash
}

func (transfer *Transfer) GetInputs() []*transferoutput.Reference {
	return transfer.inputs
}

func (transfer *Transfer) AddInput(input *transferoutput.Reference) *Transfer {
	transfer.inputs = append(transfer.inputs, input)

	return transfer
}

func (transfer *Transfer) AddOutput(address address.Address, balance *ColoredBalance) *Transfer {
	if _, addressExists := transfer.outputs[address]; !addressExists {
		transfer.outputs[address] = make([]*ColoredBalance, 0)
	}

	transfer.outputs[address] = append(transfer.outputs[address], balance)

	return transfer
}

func (transfer *Transfer) GetOutputs() map[address.Address][]*ColoredBalance {
	return transfer.outputs
}

func (transfer *Transfer) MarshalBinary() (data []byte, err error) {
	return
}

func (transfer *Transfer) UnmarshalBinary(data []byte) (err error) {
	return
}
