package v1

import (
	"github.com/iotaledger/goshimmer/packages/errors"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/interfaces"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/reality"
)

type LedgerState struct {
	storageId       []byte
	transferOutputs interfaces.TransferOutputStorage
	realities       interfaces.RealityStorage
	options         *LedgerStateOptions
}

func NewLedgerState(storageId []byte, options ...LedgerStateOption) (result interfaces.LedgerState) {
	ledgerStateOptions := DEFAULT_LEDGER_STATE_OPTIONS.Override(options...)

	realityStorage := ledgerStateOptions.RealityStorageFactory(storageId)

	result = &LedgerState{
		storageId:       storageId,
		options:         ledgerStateOptions,
		transferOutputs: ledgerStateOptions.TransferOutputStorageFactory(storageId),
		realities:       realityStorage,
	}

	newReality := reality.New(result, MAIN_REALITY_ID)
	if storeErr := realityStorage.StoreReality(newReality); storeErr != nil {
		panic(storeErr)
	}

	return
}

func (ledgerState *LedgerState) AddTransferOutput(transferOutput interfaces.TransferOutput) interfaces.LedgerState {
	if err := ledgerState.transferOutputs.StoreTransferOutput(transferOutput); err != nil {
		panic(err)
	}

	return ledgerState
}

func (ledgerState *LedgerState) GetTransferOutput(transferOutputReference interfaces.TransferOutputReference) interfaces.TransferOutput {
	if transferOutput, err := ledgerState.transferOutputs.LoadTransferOutput(transferOutputReference); err != nil {
		panic(err)
	} else {
		return transferOutput
	}
}

func (ledgerState *LedgerState) ForEachTransferOutput(callback func(transferOutput interfaces.TransferOutput), filter ...interfaces.TransferOutputStorageFilter) {
	ledgerState.transferOutputs.ForEach(callback, filter...)
}

func (ledgerState *LedgerState) BookTransfer(transfer interfaces.Transfer) errors.IdentifiableError {
	if !transfer.IsValid(ledgerState) {
		return ErrInvalidTransfer.Derive("balance of transfer is invalid")
	}

	realities := make([]hash.Reality, 0)
	for _, input := range transfer.GetInputs() {
		transferOutput := ledgerState.GetTransferOutput(input)
		if transferOutput == nil {
			return ErrInvalidTransfer.Derive("referenced transfer output doesn't exist")
		}

		realities = append(realities, transferOutput.GetRealityId())
	}

	aggregatedReality := ledgerState.MergeRealities(realities...)
	if err := ledgerState.realities.StoreReality(aggregatedReality); err != nil {
		return err
	}

	aggregatedReality.BookTransfer(transfer)
	// determine the transferoutputs
	// aggregate their realities
	// persist reality
	// book funds in this reality

	return nil
}

func (ledgerState *LedgerState) CreateReality(realityId hash.Reality) interfaces.Reality {
	if loadedReality, err := ledgerState.realities.LoadReality(realityId); err != nil {
		panic(err)
	} else if loadedReality != nil {
		return loadedReality
	} else {
		newReality := reality.New(ledgerState, realityId, MAIN_REALITY_ID)

		if storeErr := ledgerState.realities.StoreReality(newReality); storeErr != nil {
			panic(storeErr)
		}

		return newReality
	}
}

func (ledgerState *LedgerState) GetReality(realityId hash.Reality) interfaces.Reality {
	if loadedReality, loadedRealityErr := ledgerState.realities.LoadReality(realityId); loadedRealityErr != nil {
		panic(loadedRealityErr)
	} else {
		return loadedReality
	}
}

func (ledgerState *LedgerState) MergeRealities(realityIds ...hash.Reality) interfaces.Reality {
	switch len(realityIds) {
	case 0:
		if loadedReality, loadedRealityErr := ledgerState.realities.LoadReality(MAIN_REALITY_ID); loadedRealityErr != nil {
			panic(loadedRealityErr)
		} else {
			return loadedReality
		}
	case 1:
		if loadedReality, loadedRealityErr := ledgerState.realities.LoadReality(realityIds[0]); loadedRealityErr != nil {
			panic(loadedRealityErr)
		} else {
			return loadedReality
		}
	default:
		aggregatedRealities := make(map[hash.Reality]interfaces.Reality)

		for _, realityId := range realityIds {
			if _, exists := aggregatedRealities[realityId]; exists {
				continue
			}

			switchedRealities := make(map[hash.Reality]interfaces.Reality)
			realityIncluded := false
			for independentRealityId, independentReality := range aggregatedRealities {
				if independentReality.DescendsFromReality(realityId) {
					realityIncluded = true

					break
				} else {
					if loadedReality, loadedRealityErr := ledgerState.realities.LoadReality(realityId); loadedRealityErr != nil {
						panic(loadedRealityErr)
					} else if loadedReality == nil {
						return nil
					} else if loadedReality.DescendsFromReality(independentRealityId) {
						switchedRealities[independentRealityId] = loadedReality

						realityIncluded = true

						break
					}
				}
			}
			for oldId, newReality := range switchedRealities {
				delete(aggregatedRealities, oldId)
				aggregatedRealities[newReality.GetId()] = newReality
			}
			if realityIncluded {
				continue
			}

			if loadedReality, loadedRealityErr := ledgerState.realities.LoadReality(realityId); loadedRealityErr != nil {
				panic(loadedRealityErr)
			} else {
				aggregatedRealities[realityId] = loadedReality
			}
		}

		if len(aggregatedRealities) == 1 {
			for _, independentReality := range aggregatedRealities {
				return independentReality
			}
		} else {
			sortedRealityIds := make([]hash.Reality, len(aggregatedRealities))
			counter := 0
			for realityId := range aggregatedRealities {
				sortedRealityIds[counter] = realityId

				counter++
			}
			newReality := reality.New(ledgerState, "AGGREGATED", sortedRealityIds...)

			if storeErr := ledgerState.realities.StoreReality(newReality); storeErr != nil {
				panic(storeErr)
			}

			return newReality
		}

		return nil
	}
}

func (ledgerState *LedgerState) ForEachReality(callback func(reality interfaces.Reality), filter ...interfaces.TransferOutputStorageFilter) {
}
