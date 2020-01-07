package valuetangle

import (
	"container/list"

	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/tangle"
	"github.com/iotaledger/goshimmer/packages/binary/tangle/approvers"
	"github.com/iotaledger/goshimmer/packages/binary/tangle/missingtransaction"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/valuetransfer"
	"github.com/iotaledger/goshimmer/packages/binary/transactionmetadata"
	"github.com/iotaledger/goshimmer/packages/binary/types"
	"github.com/iotaledger/goshimmer/packages/binary/valuetangle/model"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
	"github.com/iotaledger/hive.go/async"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/objectstorage"
)

// The value tangle defines an "ontology" on top of the tangle that "sees"" the value transfers as a hidden tangle in
// the tangle.
type ValueTangle struct {
	tangle *tangle.Tangle

	transferMetadataStorage *objectstorage.ObjectStorage
	consumersStorage        *objectstorage.ObjectStorage
	missingTransferStorage  *objectstorage.ObjectStorage

	storeTransactionsWorkerPool async.WorkerPool
	solidifierWorkerPool        async.WorkerPool
	// cleanupWorkerPool           async.WorkerPool

	Events Events
}

func New(tangle *tangle.Tangle) (valueTangle *ValueTangle) {
	valueTangle = &ValueTangle{
		tangle: tangle,
	}

	tangle.Events.TransactionSolid.Attach(events.NewClosure(valueTangle.attachTransaction))
	tangle.Events.TransactionRemoved.Attach(events.NewClosure(valueTangle.deleteTransfer))

	return
}

func (valueTangle *ValueTangle) attachTransaction(cachedTransaction *transaction.CachedTransaction, cachedTransactionMetadata *transactionmetadata.CachedTransactionMetadata) {
	valueTangle.storeTransactionsWorkerPool.Submit(func() {
		valueTangle.storeTransactionWorker(cachedTransaction, cachedTransactionMetadata)
	})
}

// Retrieves a transaction from the tangle.
func (valueTangle *ValueTangle) GetTransfer(transactionId transaction.Id) *model.CachedValueTransfer {
	cachedTransaction := valueTangle.tangle.GetTransaction(transactionId)

	// return an empty result if the transaction is no value transaction
	if tx := cachedTransaction.Unwrap(); tx != nil && tx.GetPayload().GetType() != valuetransfer.Type {
		cachedTransaction.Release()

		return &model.CachedValueTransfer{CachedTransaction: &transaction.CachedTransaction{CachedObject: objectstorage.NewEmptyCachedObject(transactionId[:])}}
	}

	return &model.CachedValueTransfer{CachedTransaction: cachedTransaction}
}

// Retrieves the metadata of a transaction from the tangle.
func (valueTangle *ValueTangle) GetTransferMetadata(transactionId transaction.Id) *transactionmetadata.CachedTransactionMetadata {
	return &transactionmetadata.CachedTransactionMetadata{CachedObject: valueTangle.transferMetadataStorage.Load(transactionId[:])}
}

func (valueTangle *ValueTangle) deleteTransfer(transactionId transaction.Id) {

}

// Marks the tangle as stopped, so it will not accept any new transactions (waits for all backgroundTasks to finish.
func (valueTangle *ValueTangle) Shutdown() *ValueTangle {
	valueTangle.storeTransactionsWorkerPool.ShutdownGracefully()

	return valueTangle
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

	valueTransfer, transferId := cachedValueTransfer.Unwrap()
	if valueTransfer == nil {
		cachedValueTransfer.Release()

		return
	}

	for transferId := range valueTangle.getInputsMap(valueTransfer) {
		addTransferToConsumers(transferId, transferId)
	}

	if valueTangle.missingTransferStorage.DeleteIfPresent(transferId[:]) {
		valueTangle.Events.MissingTransferReceived.Trigger(transferId)
	}

	valueTangle.Events.TransferAttached.Trigger(cachedValueTransfer)

	valueTangle.solidifierWorkerPool.Submit(func() {
		valueTangle.solidifyTransferWorker(cachedValueTransfer, transferId)
	})
}

// Worker that solidifies the transactions (recursively from past to present).
func (valueTangle *ValueTangle) solidifyTransferWorker(cachedValueTransfer *model.CachedValueTransfer, transferId transfer.Id) {
	isTransferMarkedAsSolid := func(transactionId transaction.Id) bool {
		if transactionId == transaction.EmptyId {
			return true
		}

		transferMetadataCached := valueTangle.GetTransferMetadata(transactionId)
		if transactionMetadata := transferMetadataCached.Unwrap(); transactionMetadata == nil {
			transferMetadataCached.Release()

			// if transaction is missing and was not reported as missing, yet
			if cachedMissingTransfer, missingTransactionStored := valueTangle.missingTransferStorage.StoreIfAbsent(transactionId[:], missingtransaction.New(transactionId)); missingTransactionStored {
				cachedMissingTransfer.Consume(func(object objectstorage.StorableObject) {
					valueTangle.monitorMissingTransactionWorker(object.(*missingtransaction.MissingTransaction).GetTransactionId())
				})
			}

			return false
		} else if !transactionMetadata.IsSolid() {
			transferMetadataCached.Release()

			return false
		}
		transferMetadataCached.Release()

		return true
	}

	isTransactionSolid := func(transaction *transaction.Transaction, transactionMetadata *transactionmetadata.TransactionMetadata) bool {
		if transaction == nil || transaction.IsDeleted() {
			return false
		}

		if transactionMetadata == nil || transactionMetadata.IsDeleted() {
			return false
		}

		if transactionMetadata.IsSolid() {
			return true
		}

		// 1. check tangle solidity
		isTrunkSolid := isTransferMarkedAsSolid(transaction.GetTrunkTransactionId())
		isBranchSolid := isTransferMarkedAsSolid(transaction.GetBranchTransactionId())
		if isTrunkSolid && isBranchSolid {
			// 2. check payload solidity
			return true
		}

		return false
	}

	popElementsFromStack := func(stack *list.List) (*transaction.CachedTransaction, *transactionmetadata.CachedTransactionMetadata) {
		currentSolidificationEntry := stack.Front()
		currentCachedTransaction := currentSolidificationEntry.Value.([2]interface{})[0]
		currentCachedTransactionMetadata := currentSolidificationEntry.Value.([2]interface{})[1]
		stack.Remove(currentSolidificationEntry)

		return currentCachedTransaction.(*transaction.CachedTransaction), currentCachedTransactionMetadata.(*transactionmetadata.CachedTransactionMetadata)
	}

	// initialize the stack
	solidificationStack := list.New()
	solidificationStack.PushBack([2]interface{}{cachedValueTransfer, transferId})

	// process transactions that are supposed to be checked for solidity recursively
	for solidificationStack.Len() > 0 {
		currentCachedTransaction, currentCachedTransactionMetadata := popElementsFromStack(solidificationStack)

		currentTransaction := currentCachedTransaction.Unwrap()
		currentTransactionMetadata := currentCachedTransactionMetadata.Unwrap()
		if currentTransaction == nil || currentTransactionMetadata == nil {
			currentCachedTransaction.Release()
			currentCachedTransactionMetadata.Release()

			continue
		}

		// if current transaction is solid and was not marked as solid before: mark as solid and propagate
		if isTransactionSolid(currentTransaction, currentTransactionMetadata) && currentTransactionMetadata.SetSolid(true) {
			valueTangle.Events.TransferSolid.Trigger(currentCachedTransaction, currentCachedTransactionMetadata)

			valueTangle.GetConsumers(currentTransaction.GetId()).Consume(func(object objectstorage.StorableObject) {
				for approverTransactionId := range object.(*approvers.Approvers).Get() {
					solidificationStack.PushBack([2]interface{}{
						valueTangle.GetTransfer(approverTransactionId),
						valueTangle.GetTransferMetadata(approverTransactionId),
					})
				}
			})
		}

		// release cached results
		currentCachedTransaction.Release()
		currentCachedTransactionMetadata.Release()
	}
}

func (valueTangle *ValueTangle) getInputsMap(valueTransfer *valuetransfer.ValueTransfer) (result map[transfer.Id]map[address.Address]types.Empty) {
	result = make(map[transfer.Id]map[address.Address]types.Empty)

	for _, transferOutputReference := range valueTransfer.GetInputs() {
		addressMap, addressMapExists := result[transferOutputReference.GetTransferHash()]
		if !addressMapExists {
			addressMap = make(map[address.Address]types.Empty)

			result[transferOutputReference.GetTransferHash()] = addressMap
		}

		addressMap[transferOutputReference.GetAddress()] = types.Void
	}

	return
}
