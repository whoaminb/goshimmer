package ledgerstate

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/errors"
)

type LedgerState struct {
	mainReality  *Reality
	subRealities map[string]*Reality
}

func NewLedgerState() *LedgerState {
	return &LedgerState{
		mainReality:  NewReality(),
		subRealities: make(map[string]*Reality),
	}
}

func (ledgerState *LedgerState) GetReality() *Reality {
	return ledgerState.mainReality
}

func (ledgerState *LedgerState) verifyTransferBalances(transfer *Transfer) bool {
	totalColoredBalances := make(map[ColorHash]uint64)

	for addressHash, transferOutputReferences := range transfer.inputs {
		unspentTransferOutputs := ledgerState.GetUnspentTransferOutputs(addressHash)

		// process inputs
		for _, transferOutputReference := range transferOutputReferences {
			if transferOutput := unspentTransferOutputs[transferOutputReference.transferHash]; transferOutput.Exists() {
				for colorHash, coloredBalance := range transferOutput.coloredBalances {
					totalColoredBalances[colorHash] += coloredBalance.balance
				}
			}
		}

		// process outputs
		for _, transferOutput := range transfer.outputs {
			for colorHash, coloredBalance := range transferOutput {
				totalColoredBalances[colorHash] -= coloredBalance.balance

				if totalColoredBalances[colorHash] == 0 {
					delete(totalColoredBalances, colorHash)
				}
			}
		}
	}

	// check if transfer is valid (sum of funds is 0)
	return len(totalColoredBalances) == 0
}

func (ledgerState *LedgerState) BookTransfer(transfer *Transfer) (err errors.IdentifiableError) {
	if !ledgerState.verifyTransferBalances(transfer) {
		err = ErrInvalidTransfer.Derive("balance of transfer is invalid")
	}

	for addressHash, transferOutputReferences := range transfer.inputs {
		unspentTransferOutputs := ledgerState.GetUnspentTransferOutputs(addressHash)

		// process inputs
		for _, transferOutputReference := range transferOutputReferences {
			if transferOutput := unspentTransferOutputs[transferOutputReference.transferHash]; transferOutput.Exists() {
				//transferOutput.MarkAsConsumed(transfer.hash)
			} else {
				// ERROR
			}
		}

		fmt.Println("AA", unspentTransferOutputs)

		// process outputs
		for _, transferOutput := range transfer.outputs {
			for colorHash, coloredBalance := range transferOutput {
				// create output
				fmt.Println("OI", colorHash, coloredBalance)
			}
		}
	}

	return
}

func (ledgerState *LedgerState) AddAddress(address *Address) *LedgerState {
	ledgerState.mainReality.SetAddress(address)

	return ledgerState
}

func (ledgerState *LedgerState) GetUnspentTransferOutputs(address AddressHash, includedSubRealities ...TransferHash) (result TransferOutputs) {
	result = make(TransferOutputs)

	for _, reality := range ledgerState.GetRealities(includedSubRealities...) {
		if address := reality.GetAddress(address); address.Exists() {
			for transferHash, coloredBalance := range address.GetUnspentTransferOutputs() {
				result[transferHash] = coloredBalance
			}
		}
	}

	return
}

func (ledgerState *LedgerState) GetRealities(includedSubRealities ...TransferHash) (realities []*Reality) {
	realities = append(realities, ledgerState.mainReality)

	for _, subRealityToInclude := range includedSubRealities {
		if subReality := ledgerState.subRealities[subRealityToInclude]; subReality.Exists() {
			realities = append(realities, subReality)
		}
	}

	return
}
