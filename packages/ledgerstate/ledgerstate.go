package ledgerstate

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/objectstorage"
)

type LedgerState struct {
	storageId       []byte
	transferOutputs *objectstorage.ObjectStorage
	realities       *objectstorage.ObjectStorage
}

func NewLedgerState(storageId string) *LedgerState {
	return &LedgerState{
		storageId:       []byte(storageId),
		transferOutputs: objectstorage.New(storageId+"TRANSFER_OUTPUTS", &TransferOutput{}),
		realities:       objectstorage.New(storageId+"REALITIES", &Reality{}),
	}
}

func (ledgerState *LedgerState) AddTransferOutput(transferHash TransferHash, addressHash AddressHash, balances ...*ColoredBalance) *LedgerState {
	ledgerState.storeTransferOutput(NewTransferOutput(ledgerState, MAIN_REALITY_ID, transferHash, addressHash, balances...)).Release()

	return ledgerState
}

func (ledgerState *LedgerState) storeTransferOutput(transferOutput *TransferOutput) *objectstorage.CachedObject {
	return ledgerState.transferOutputs.Store(transferOutput)
}

func (ledgerState *LedgerState) GetTransferOutput(transferOutputReference *TransferOutputReference) *objectstorage.CachedObject {
	if cachedTransferOutput, err := ledgerState.transferOutputs.Load(transferOutputReference.GetStorageKey()); err != nil {
		panic(err)
	} else {
		if cachedTransferOutput.Exists() {
			if transferOutput := cachedTransferOutput.Get().(*TransferOutput); transferOutput != nil {
				transferOutput.ledgerState = ledgerState
			}
		}

		return cachedTransferOutput
	}
}

func (ledgerState *LedgerState) CreateReality(id RealityId) {
	ledgerState.realities.Store(newReality(id, MAIN_REALITY_ID))
}

func (ledgerState *LedgerState) GetReality(id RealityId) *objectstorage.CachedObject {
	if cachedObject, err := ledgerState.realities.Load(id[:]); err != nil {
		panic(err)
	} else {
		if cachedObject.Exists() {
			if reality := cachedObject.Get().(*Reality); reality != nil {
				reality.ledgerState = ledgerState
			}
		}

		return cachedObject
	}
}

func (ledgerState *LedgerState) Prune() *LedgerState {
	fmt.Println("PRUNED DATABASE")

	return ledgerState
}
