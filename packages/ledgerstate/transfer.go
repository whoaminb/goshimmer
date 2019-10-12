package ledgerstate

type Transfer struct {
	hash    TransferHash
	inputs  map[AddressHash]map[TransferHash]*TransferOutputReference
	outputs map[AddressHash]map[ColorHash]*ColoredBalance
}

func NewTransfer(hash TransferHash) *Transfer {
	return &Transfer{
		hash:    hash,
		inputs:  make(map[AddressHash]map[TransferHash]*TransferOutputReference),
		outputs: make(map[AddressHash]map[ColorHash]*ColoredBalance),
	}
}

func (transfer *Transfer) AddInputs(address AddressHash, transferOutputs ...TransferHash) *Transfer {
	if len(transferOutputs) >= 1 {
		addressEntry, addressExists := transfer.inputs[address]
		if !addressExists {
			addressEntry = make(map[TransferHash]*TransferOutputReference)

			transfer.inputs[address] = addressEntry
		}

		for _, transferOutput := range transferOutputs {
			addressEntry[transferOutput] = NewTransferOutputReference(address, transferOutput)
		}
	}

	return transfer
}

func (transfer *Transfer) AddOutput(address AddressHash, balance *ColoredBalance) *Transfer {
	addressEntry, addressExists := transfer.outputs[address]
	if !addressExists {
		addressEntry = make(map[ColorHash]*ColoredBalance)

		transfer.outputs[address] = addressEntry
	}

	addressEntry[balance.color] = balance

	return transfer
}
