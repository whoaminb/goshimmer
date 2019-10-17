package interfaces

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
)

type Transfer interface {
	GetHash() hash.Transfer
	IsValid(ledgerState LedgerState) bool
	AddInput(input TransferOutputReference) Transfer
	GetInputs() []TransferOutputReference
	GetOutputs() map[hash.Address]map[hash.Color]ColoredBalance
	AddOutput(address hash.Address, balance ColoredBalance) Transfer
}
