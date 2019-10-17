package address

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/interfaces"
)

type Address struct {
	ledgerState interfaces.LedgerState
	realityId   hash.Reality
	hash        hash.Address
}

func New(ledgerState interfaces.LedgerState, realityId hash.Reality, hash hash.Address) interfaces.RealityAddress {
	return &Address{
		ledgerState: ledgerState,
		realityId:   realityId,
		hash:        hash,
	}
}

func (address *Address) GetHash() hash.Address {
	return address.hash
}

func (address *Address) GetRealityId() hash.Reality {
	return address.realityId
}

func (address *Address) GetReality() interfaces.Reality {
	return address.ledgerState.GetReality(address.realityId)
}

func (address *Address) GetUnspentTransferOutputs() (unspentTransferOutputs []interfaces.TransferOutput) {
	unspentTransferOutputs = make([]interfaces.TransferOutput, 0)

	address.collectUnspentTransferOutputs(address.realityId, &unspentTransferOutputs)
	for ancestorRealityId := range address.GetReality().GetAncestorRealities() {
		address.collectUnspentTransferOutputs(ancestorRealityId, &unspentTransferOutputs)
	}

	return unspentTransferOutputs
}

func (address *Address) collectUnspentTransferOutputs(realityId hash.Reality, unspentTransferOutputs *[]interfaces.TransferOutput) {
	address.ledgerState.ForEachTransferOutput(func(transferOutput interfaces.TransferOutput) {
		*unspentTransferOutputs = append(*unspentTransferOutputs, transferOutput)
	}, interfaces.FilterRealities(realityId), interfaces.FilterAddresses(address.hash), interfaces.FilterUnspent())
}

func (address *Address) GetTransferOutputs() map[hash.Transfer]interfaces.TransferOutput {
	return nil
}

func (address *Address) GetBalances() map[hash.Color]uint64 {
	balances := make(map[hash.Color]uint64)

	for _, unspentTransferOutput := range address.GetUnspentTransferOutputs() {
		for colorHash, balance := range unspentTransferOutput.GetColoredBalances() {
			balances[colorHash] += balance.GetValue()
		}
	}

	return balances
}
