package reality

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/transferoutput"

	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/address"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/interfaces"
)

type Reality struct {
	ledgerState      interfaces.LedgerState
	id               hash.Reality
	parentRealityIds []hash.Reality
	parentRealities  map[hash.Reality]interfaces.Reality
}

func New(ledgerState interfaces.LedgerState, id hash.Reality, parentRealityIds ...hash.Reality) interfaces.Reality {
	return &Reality{
		ledgerState:      ledgerState,
		id:               id,
		parentRealityIds: parentRealityIds,
	}
}

func (reality *Reality) SetLedgerState(ledgerState interfaces.LedgerState) {
	reality.ledgerState = ledgerState
}

func (reality *Reality) GetId() hash.Reality {
	return reality.id
}

func (reality *Reality) GetParentRealityIds() []hash.Reality {
	return reality.parentRealityIds
}

func (reality *Reality) GetAddress(addressHash hash.Address) interfaces.RealityAddress {
	return address.New(reality.ledgerState, reality.id, addressHash)
}

func (reality *Reality) GetParentRealities() map[hash.Reality]interfaces.Reality {
	if reality.parentRealities == nil {
		parentRealities := make(map[hash.Reality]interfaces.Reality)
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

func (reality *Reality) GetAncestorRealities() (result map[hash.Reality]interfaces.Reality) {
	result = make(map[hash.Reality]interfaces.Reality, 1)

	for _, parentReality := range reality.GetParentRealities() {
		result[parentReality.GetId()] = reality

		for _, ancestor := range parentReality.GetAncestorRealities() {
			result[ancestor.GetId()] = ancestor
		}
	}

	return
}

func (reality *Reality) DescendsFromReality(realityId hash.Reality) bool {
	if reality.id == realityId {
		return true
	} else {
		_, exists := reality.GetAncestorRealities()[realityId]

		return exists
	}
}

func (reality *Reality) BookTransfer(transfer interfaces.Transfer) errors.IdentifiableError {
	// process outputs
	for addressHash, transferOutput := range transfer.GetOutputs() {
		for _, coloredBalance := range transferOutput {
			createdTransferOutput := transferoutput.New(reality.ledgerState, reality.id, addressHash, transfer.GetHash(), coloredBalance)
			reality.ledgerState.AddTransferOutput(createdTransferOutput)
		}
	}

	return nil
}

func (reality *Reality) Exists() bool {
	return reality != nil
}
