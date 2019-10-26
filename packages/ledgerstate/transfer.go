package ledgerstate

type Transfer struct {
	hash    TransferHash
	inputs  []*TransferOutputReference
	outputs map[AddressHash][]*ColoredBalance
}

func NewTransfer(transferHash TransferHash) *Transfer {
	return &Transfer{
		hash:    transferHash,
		inputs:  make([]*TransferOutputReference, 0),
		outputs: make(map[AddressHash][]*ColoredBalance),
	}
}

func (transfer *Transfer) GetHash() TransferHash {
	return transfer.hash
}

func (transfer *Transfer) GetInputs() []*TransferOutputReference {
	return transfer.inputs
}

func (transfer *Transfer) AddInput(input *TransferOutputReference) *Transfer {
	transfer.inputs = append(transfer.inputs, input)

	return transfer
}

func (transfer *Transfer) AddOutput(address AddressHash, balance *ColoredBalance) *Transfer {
	if _, addressExists := transfer.outputs[address]; !addressExists {
		transfer.outputs[address] = make([]*ColoredBalance, 0)
	}

	transfer.outputs[address] = append(transfer.outputs[address], balance)

	return transfer
}

func (transfer *Transfer) GetOutputs() map[AddressHash][]*ColoredBalance {
	return transfer.outputs
}
