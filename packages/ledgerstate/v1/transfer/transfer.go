package transfer

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/interfaces"
)

type Transfer struct {
	hash    hash.Transfer
	inputs  []interfaces.TransferOutputReference
	outputs map[hash.Address]map[hash.Color]interfaces.ColoredBalance
}

func New(transferHash hash.Transfer) interfaces.Transfer {
	return &Transfer{
		hash:    transferHash,
		inputs:  make([]interfaces.TransferOutputReference, 0),
		outputs: make(map[hash.Address]map[hash.Color]interfaces.ColoredBalance),
	}
}

func (transfer *Transfer) GetHash() hash.Transfer {
	return transfer.hash
}

func (transfer *Transfer) IsValid(ledgerState interfaces.LedgerState) bool {
	totalColoredBalances := make(map[hash.Color]uint64)

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

func (transfer *Transfer) AddInput(input interfaces.TransferOutputReference) interfaces.Transfer {
	transfer.inputs = append(transfer.inputs, input)

	return transfer
}

func (transfer *Transfer) GetInputs() []interfaces.TransferOutputReference {
	return transfer.inputs
}

func (transfer *Transfer) AddOutput(address hash.Address, balance interfaces.ColoredBalance) interfaces.Transfer {
	addressEntry, addressExists := transfer.outputs[address]
	if !addressExists {
		addressEntry = make(map[hash.Color]interfaces.ColoredBalance)

		transfer.outputs[address] = addressEntry
	}

	addressEntry[balance.GetColor()] = balance

	return transfer
}

func (transfer *Transfer) GetOutputs() map[hash.Address]map[hash.Color]interfaces.ColoredBalance {
	return transfer.outputs
}
