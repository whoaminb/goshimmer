package ledgerstate

import (
	"github.com/iotaledger/goshimmer/packages/errors"
)

type RealityId string

type Reality struct {
	ledgerState      *LedgerState
	id               RealityId
	parentRealityIds []RealityId
	parentRealities  map[RealityId]*Reality
}

func newReality(ledgerState *LedgerState, id RealityId, parentRealityIds ...RealityId) *Reality {
	return &Reality{
		ledgerState:      ledgerState,
		id:               id,
		parentRealityIds: parentRealityIds,
	}
}

func (reality *Reality) SetLedgerState(ledgerState *LedgerState) {
	reality.ledgerState = ledgerState
}

func (reality *Reality) GetId() RealityId {
	return reality.id
}

func (reality *Reality) GetParentRealityIds() []RealityId {
	return reality.parentRealityIds
}

func (reality *Reality) GetAddress(addressHash AddressHash) *Address {
	return NewAddress(reality.ledgerState, reality.id, addressHash)
}

func (reality *Reality) GetParentRealities() map[RealityId]*Reality {
	if reality.parentRealities == nil {
		parentRealities := make(map[RealityId]*Reality)
		for _, parentRealityId := range reality.parentRealityIds {
			if loadedParentReality := reality.ledgerState.GetReality(parentRealityId); loadedParentReality == nil {
				panic("could not load parent reality " + parentRealityId)
			} else {
				parentRealities[loadedParentReality.GetId()] = loadedParentReality
			}
		}
		reality.parentRealities = parentRealities
	}

	return reality.parentRealities
}

func (reality *Reality) GetAncestorRealities() (result map[RealityId]*Reality) {
	result = make(map[RealityId]*Reality, 1)

	for _, parentReality := range reality.GetParentRealities() {
		result[parentReality.GetId()] = reality

		for _, ancestor := range parentReality.GetAncestorRealities() {
			result[ancestor.GetId()] = ancestor
		}
	}

	return
}

func (reality *Reality) DescendsFromReality(realityId RealityId) bool {
	if reality.id == realityId {
		return true
	} else {
		_, exists := reality.GetAncestorRealities()[realityId]

		return exists
	}
}

func (reality *Reality) BookTransfer(transfer *Transfer) errors.IdentifiableError {
	// process outputs
	for addressHash, transferOutput := range transfer.GetOutputs() {
		for _, coloredBalance := range transferOutput {
			createdTransferOutput := NewTransferOutput(reality.ledgerState, reality.id, addressHash, transfer.GetHash(), coloredBalance)
			reality.ledgerState.AddTransferOutput(createdTransferOutput)
		}
	}

	return nil
}

func (reality *Reality) Exists() bool {
	return reality != nil
}
