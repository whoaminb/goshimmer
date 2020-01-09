package valuetangle

import (
	"container/list"
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/packages/binary/address"

	"github.com/iotaledger/goshimmer/packages/binary/types"

	"github.com/iotaledger/goshimmer/packages/storageprefix"

	"github.com/iotaledger/goshimmer/packages/binary/tangle"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/valuetransfer"
	"github.com/iotaledger/goshimmer/packages/binary/transactionmetadata"
	"github.com/iotaledger/goshimmer/packages/binary/valuetangle/model"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
	"github.com/iotaledger/hive.go/async"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/objectstorage"
)

const (
	MaxMissingTimeBeforeCleanup = 30 * time.Second
	MissingCheckInterval        = 5 * time.Second
)

// The "value tangle" defines an "ontology" on top of the tangle that "sees"" the value transfers as a hidden tangle in
// the tangle.
type ValueTangle struct {
	tangle *tangle.Tangle

	transferMetadataStorage *objectstorage.ObjectStorage
	consumersStorage        *objectstorage.ObjectStorage
	missingTransferStorage  *objectstorage.ObjectStorage

	storeTransactionsWorkerPool async.WorkerPool
	solidifierWorkerPool        async.WorkerPool
	cleanupWorkerPool           async.WorkerPool

	Events Events
}

func New(tangle *tangle.Tangle) (valueTangle *ValueTangle) {
	valueTangle = &ValueTangle{
		tangle: tangle,

		transferMetadataStorage: objectstorage.New(append(tangle.GetStorageId(), storageprefix.ValueTangleTransferMetadata...), model.TransferMetadataFromStorage),
		consumersStorage:        objectstorage.New(append(tangle.GetStorageId(), storageprefix.ValueTangleConsumers...), model.ConsumersFromStorage),
		missingTransferStorage:  objectstorage.New(append(tangle.GetStorageId(), storageprefix.TangleMissingTransaction...), model.MissingTransferFromStorage),

		Events: *newEvents(),
	}

	valueTangle.solidifierWorkerPool.Tune(1024)

	tangle.Events.TransactionSolid.Attach(events.NewClosure(valueTangle.attachTransaction))
	tangle.Events.TransactionRemoved.Attach(events.NewClosure(valueTangle.deleteTransfer))

	return
}

func (valueTangle *ValueTangle) attachTransaction(cachedTransaction *transaction.CachedTransaction, cachedTransactionMetadata *transactionmetadata.CachedTransactionMetadata) {
	valueTangle.storeTransactionsWorkerPool.Submit(func() {
		valueTangle.storeTransactionWorker(cachedTransaction, cachedTransactionMetadata)
	})
}

// Retrieves a transfer from the tangle.
func (valueTangle *ValueTangle) GetTransfer(transferId transfer.Id) *model.CachedValueTransfer {
	cachedTransaction := valueTangle.tangle.GetTransaction(transaction.NewId(transferId[:]))

	// return an empty result if the transfer is no value transfer
	if tx := cachedTransaction.Unwrap(); tx != nil && tx.GetPayload().GetType() != valuetransfer.Type {
		cachedTransaction.Release()

		return &model.CachedValueTransfer{CachedTransaction: &transaction.CachedTransaction{CachedObject: objectstorage.NewEmptyCachedObject(transferId[:])}}
	}

	return &model.CachedValueTransfer{CachedTransaction: cachedTransaction}
}

// Retrieves the metadata of a transfer from the tangle.
func (valueTangle *ValueTangle) GetTransferMetadata(transferId transfer.Id) *model.CachedTransferMetadata {
	return &model.CachedTransferMetadata{CachedObject: valueTangle.transferMetadataStorage.Load(transferId[:])}
}

// Retrieves the approvers of a transfer from the tangle.
func (valueTangle *ValueTangle) GetConsumers(transferId transfer.Id) *model.CachedConsumers {
	return &model.CachedConsumers{CachedObject: valueTangle.consumersStorage.Load(transferId[:])}
}

func (valueTangle *ValueTangle) deleteTransfer(transferIf transfer.Id) {

}

// Marks the tangle as stopped, so it will not accept any new transfers (waits for all backgroundTasks to finish.
func (valueTangle *ValueTangle) Shutdown() *ValueTangle {
	valueTangle.tangle.Shutdown()

	valueTangle.storeTransactionsWorkerPool.ShutdownGracefully()
	valueTangle.solidifierWorkerPool.ShutdownGracefully()
	valueTangle.cleanupWorkerPool.ShutdownGracefully()

	return valueTangle
}

// Resets the database and deletes all objects (good for testing or "node resets").
func (valueTangle *ValueTangle) Prune() error {
	for _, storage := range []*objectstorage.ObjectStorage{
		valueTangle.transferMetadataStorage,
		valueTangle.consumersStorage,
		valueTangle.missingTransferStorage,
	} {
		if err := storage.Prune(); err != nil {
			return err
		}
	}

	return nil
}

func (valueTangle *ValueTangle) storeTransactionWorker(cachedTx *transaction.CachedTransaction, cachedTxMetadata *transactionmetadata.CachedTransactionMetadata) {
	addTransferToConsumers := func(transferId transfer.Id, consumedTransferHash transfer.Id) {
		cachedConsumers := valueTangle.consumersStorage.ComputeIfAbsent(consumedTransferHash[:], func([]byte) objectstorage.StorableObject {
			result := model.NewConsumers(consumedTransferHash)

			result.SetModified()

			return result
		})

		if _tmp := cachedConsumers.Get(); _tmp != nil {
			if consumersObject := _tmp.(*model.Consumers); consumersObject != nil {
				consumersObject.Add(transferId)

				// if the approvers got "cleaned up" while being in cache, we make sure the object gets persisted again
				consumersObject.Persist()
			}
		}

		cachedConsumers.Release()
	}

	cachedTxMetadata.Release()

	cachedValueTransfer := &model.CachedValueTransfer{CachedTransaction: cachedTx}

	valueTransfer := cachedValueTransfer.Unwrap()
	if valueTransfer == nil {
		cachedValueTransfer.Release()

		return
	}

	transferId := valueTransfer.GetId()

	cachedTransferMetadata := &model.CachedTransferMetadata{CachedObject: valueTangle.transferMetadataStorage.Store(model.NewTransferMetadata(transferId))}
	for referencedTransferId := range valueTransfer.GetInputs() {
		addTransferToConsumers(transferId, referencedTransferId)
	}

	if valueTangle.missingTransferStorage.DeleteIfPresent(transferId[:]) {
		valueTangle.Events.MissingTransferReceived.Trigger(transferId)
	}

	valueTangle.Events.TransferAttached.Trigger(cachedValueTransfer, cachedTransferMetadata)

	valueTangle.solidifierWorkerPool.Submit(func() {
		valueTangle.solidifyTransferWorker(cachedValueTransfer, cachedTransferMetadata)
	})
}

// Worker that solidifies the transfers (recursively from past to present).
func (valueTangle *ValueTangle) solidifyTransferWorker(cachedValueTransfer *model.CachedValueTransfer, cachedTransferMetadata *model.CachedTransferMetadata) {
	areInputsSolid := func(transferId transfer.Id, addresses map[address.Address]types.Empty) bool {
		transferMetadataCached := valueTangle.GetTransferMetadata(transferId)
		if transferMetadata := transferMetadataCached.Unwrap(); transferMetadata == nil {
			transferMetadataCached.Release()

			// if transfer is missing and was not reported as missing, yet
			if cachedMissingTransfer, missingTransactionStored := valueTangle.missingTransferStorage.StoreIfAbsent(transferId[:], model.NewMissingTransfer(transferId)); missingTransactionStored {
				cachedMissingTransfer.Consume(func(object objectstorage.StorableObject) {
					valueTangle.monitorMissingTransactionWorker(object.(*model.MissingTransfer).GetTransferId())
				})
			}

			return false
		} else if !transferMetadata.IsSolid() {
			transferMetadataCached.Release()

			return false
		}
		transferMetadataCached.Release()

		cachedTransfer := valueTangle.GetTransfer(transferId)
		if valueTransfer := cachedTransfer.Unwrap(); valueTransfer == nil {
			cachedTransfer.Release()

			return false
		} else {
			outputs := valueTransfer.GetOutputs()
			for address := range outputs {
				if _, addressExists := outputs[address]; !addressExists {
					cachedTransfer.Release()

					fmt.Println("INVALID TX DETECTED")

					return false
				}
			}
		}
		cachedTransfer.Release()

		return true
	}

	isTransferSolid := func(transfer *model.ValueTransfer, transferMetadata *model.TransferMetadata) bool {
		if transfer == nil || transfer.IsDeleted() {
			return false
		}

		if transferMetadata == nil || transferMetadata.IsDeleted() {
			return false
		}

		if transferMetadata.IsSolid() {
			return true
		}

		// 1. check tangle solidity
		inputsSolid := true
		for inputTransferId, addresses := range transfer.GetInputs() {
			inputsSolid = inputsSolid && areInputsSolid(inputTransferId, addresses)
		}

		return inputsSolid
	}

	popElementsFromStack := func(stack *list.List) (*model.CachedValueTransfer, *model.CachedTransferMetadata) {
		currentSolidificationEntry := stack.Front()
		currentCachedTransfer := currentSolidificationEntry.Value.([2]interface{})[0]
		currentCachedTransferMetadata := currentSolidificationEntry.Value.([2]interface{})[1]
		stack.Remove(currentSolidificationEntry)

		return currentCachedTransfer.(*model.CachedValueTransfer), currentCachedTransferMetadata.(*model.CachedTransferMetadata)
	}

	// initialize the stack
	solidificationStack := list.New()
	solidificationStack.PushBack([2]interface{}{cachedValueTransfer, cachedTransferMetadata})

	// process transfers that are supposed to be checked for solidity recursively
	for solidificationStack.Len() > 0 {
		currentCachedTransfer, currentCachedTransferMetadata := popElementsFromStack(solidificationStack)

		currentTransfer := currentCachedTransfer.Unwrap()
		currentTransferMetadata := currentCachedTransferMetadata.Unwrap()
		if currentTransfer == nil || currentTransferMetadata == nil {
			currentCachedTransfer.Release()
			currentCachedTransferMetadata.Release()

			continue
		}

		// if current transfer is solid and was not marked as solid before: mark as solid and propagate
		if isTransferSolid(currentTransfer, currentTransferMetadata) && currentTransferMetadata.SetSolid(true) {
			valueTangle.Events.TransferSolid.Trigger(currentCachedTransfer, currentCachedTransferMetadata)

			valueTangle.GetConsumers(currentTransfer.GetId()).Consume(func(object objectstorage.StorableObject) {
				for approverTransactionId := range object.(*model.Consumers).Get() {
					solidificationStack.PushBack([2]interface{}{
						valueTangle.GetTransfer(approverTransactionId),
						valueTangle.GetTransferMetadata(approverTransactionId),
					})
				}
			})
		}

		// release cached results
		currentCachedTransfer.Release()
		currentCachedTransferMetadata.Release()
	}
}

// Worker that Monitors the missing transfers (by scheduling regular checks).
func (valueTangle *ValueTangle) monitorMissingTransactionWorker(transferId transfer.Id) {
	var scheduleNextMissingCheck func(transferId transfer.Id)
	scheduleNextMissingCheck = func(transferId transfer.Id) {
		time.AfterFunc(MissingCheckInterval, func() {
			valueTangle.missingTransferStorage.Load(transferId[:]).Consume(func(object objectstorage.StorableObject) {
				missingTransfer := object.(*model.MissingTransfer)

				if time.Since(missingTransfer.GetMissingSince()) >= MaxMissingTimeBeforeCleanup {
					valueTangle.cleanupWorkerPool.Submit(func() { valueTangle.cleanupWorker(missingTransfer.GetTransferId()) })
				} else {
					valueTangle.Events.TransferMissing.Trigger(transferId)

					scheduleNextMissingCheck(transferId)
				}
			})
		})
	}
	valueTangle.Events.TransferMissing.Trigger(transferId)

	scheduleNextMissingCheck(transferId)
}

// Worker that recursively cleans up the approvers of a unsolidifiable missing transfer.
func (valueTangle *ValueTangle) cleanupWorker(transferId transfer.Id) {
	cleanupStack := list.New()
	cleanupStack.PushBack(transferId)

	for cleanupStack.Len() >= 1 {
		currentStackEntry := cleanupStack.Front()
		currentTransferId := currentStackEntry.Value.(transfer.Id)
		cleanupStack.Remove(currentStackEntry)

		valueTangle.GetConsumers(currentTransferId).Consume(func(object objectstorage.StorableObject) {
			for approverTransactionId := range object.(*model.Consumers).Get() {
				valueTangle.deleteTransfer(currentTransferId)

				cleanupStack.PushBack(approverTransactionId)
			}
		})
	}
}
