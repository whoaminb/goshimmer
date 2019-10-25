package ledgerstate

type Transfer struct {
	hash    TransferHash
	inputs  []*TransferOutputReference
	outputs map[AddressHash]map[Color]*ColoredBalance
}

func NewTransfer(transferHash TransferHash) *Transfer {
	return &Transfer{
		hash:    transferHash,
		inputs:  make([]*TransferOutputReference, 0),
		outputs: make(map[AddressHash]map[Color]*ColoredBalance),
	}
}

func (transfer *Transfer) AddInput(input *TransferOutputReference) *Transfer {
	transfer.inputs = append(transfer.inputs, input)

	return transfer
}

func (transfer *Transfer) AddOutput(address AddressHash, balance *ColoredBalance) *Transfer {
	addressEntry, addressExists := transfer.outputs[address]
	if !addressExists {
		addressEntry = make(map[Color]*ColoredBalance)

		transfer.outputs[address] = addressEntry
	}

	addressEntry[balance.GetColor()] = balance

	return transfer
}
