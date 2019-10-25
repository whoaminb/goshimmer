package ledgerstate

type TransferHash [transferHashLength]byte

type Transfer struct {
	hash    TransferHash
	inputs  []*TransferOutputReferenceOld
	outputs map[AddressHash]map[Color]*ColoredBalance
}

func NewTransfer(transferHash TransferHash) *Transfer {
	return &Transfer{
		hash:    transferHash,
		inputs:  make([]*TransferOutputReferenceOld, 0),
		outputs: make(map[AddressHash]map[Color]*ColoredBalance),
	}
}

func (transfer *Transfer) GetHash() TransferHash {
	return transfer.hash
}

func (transfer *Transfer) IsValid(ledgerState *LedgerState) bool {
	totalColoredBalances := make(map[Color]uint64)

	// process inputs
	for _, transferOutputReference := range transfer.inputs {
		if transferOutput := ledgerState.GetTransferOutput(transferOutputReference); transferOutput == nil {
			return false
		} else {
			for colorHash, coloredBalance := range transferOutput.GetColoredBalances() {
				totalColoredBalances[colorHash] += coloredBalance.GetValue()
			}
		}
	}

	// process outputs
	for _, transferOutput := range transfer.outputs {
		for colorHash, coloredBalance := range transferOutput {
			totalColoredBalances[colorHash] -= coloredBalance.GetValue()

			if totalColoredBalances[colorHash] == 0 {
				delete(totalColoredBalances, colorHash)
			}
		}
	}

	// transfer is valid if sum of funds is 0
	return len(totalColoredBalances) == 0
}

func (transfer *Transfer) AddInput(input *TransferOutputReferenceOld) *Transfer {
	transfer.inputs = append(transfer.inputs, input)

	return transfer
}

func (transfer *Transfer) GetInputs() []*TransferOutputReferenceOld {
	return transfer.inputs
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

func (transfer *Transfer) GetOutputs() map[AddressHash]map[Color]*ColoredBalance {
	return transfer.outputs
}
