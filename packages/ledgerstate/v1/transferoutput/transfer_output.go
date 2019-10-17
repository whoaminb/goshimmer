package transferoutput

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/interfaces"
)

type TransferOutput struct {
	ledgerState     interfaces.LedgerState
	realityHash     hash.Reality
	addressHash     hash.Address
	transferHash    hash.Transfer
	coloredBalances map[hash.Color]interfaces.ColoredBalance
	consumers       []hash.Transfer
}

func New(ledgerState interfaces.LedgerState, realityHash hash.Reality, addressHash hash.Address, transferHash hash.Transfer, coloredBalances ...interfaces.ColoredBalance) (result *TransferOutput) {
	result = &TransferOutput{
		ledgerState:     ledgerState,
		addressHash:     addressHash,
		transferHash:    transferHash,
		realityHash:     realityHash,
		coloredBalances: make(map[hash.Color]interfaces.ColoredBalance),
		consumers:       make([]hash.Transfer, 0),
	}

	for _, balance := range coloredBalances {
		result.coloredBalances[balance.GetColor()] = balance
	}

	return
}

func (transferOutput *TransferOutput) GetRealityId() hash.Reality {
	return transferOutput.realityHash
}

func (transferOutput *TransferOutput) GetReality(realityId hash.Reality) interfaces.Reality {
	return transferOutput.ledgerState.GetReality(realityId)
}

func (transferOutput *TransferOutput) GetAddressHash() hash.Address {
	return transferOutput.addressHash
}

func (transferOutput *TransferOutput) GetTransferHash() hash.Transfer {
	return transferOutput.transferHash
}

func (transferOutput *TransferOutput) GetColoredBalances() map[hash.Color]interfaces.ColoredBalance {
	return transferOutput.coloredBalances
}

func (transferOutput *TransferOutput) GetConsumers() []hash.Transfer {
	return transferOutput.consumers
}

func (transferOutput *TransferOutput) Exists() bool {
	return transferOutput != nil
}

func (transferOutput *TransferOutput) String() (result string) {
	result = "TransferOutput {\n"
	result += "    RealityHash:  \"" + transferOutput.realityHash + "\",\n"
	result += "    AddressHash:  \"" + transferOutput.addressHash + "\",\n"
	result += "    TransferHash: \"" + transferOutput.transferHash + "\",\n"

	for _, coloredBalance := range transferOutput.coloredBalances {
		result += "    " + coloredBalance.String() + ",\n"
	}

	result += "}"

	return
}
