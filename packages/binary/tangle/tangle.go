package tangle

import (
	"container/list"
	"time"

	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/valuetransfer"

	"github.com/iotaledger/goshimmer/packages/binary/tangle/approvers"
	"github.com/iotaledger/goshimmer/packages/binary/tangle/missingtransaction"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/data"
	"github.com/iotaledger/goshimmer/packages/binary/transactionmetadata"
	"github.com/iotaledger/goshimmer/packages/storageprefix"
	"github.com/iotaledger/hive.go/async"
	"github.com/iotaledger/hive.go/objectstorage"
)

const (
	MAX_MISSING_TIME_BEFORE_CLEANUP = 30 * time.Second
	MISSING_CHECK_INTERVAL          = 5 * time.Second
)

type Tangle struct {
	storageId []byte

	transactionStorage         *objectstorage.ObjectStorage
	transactionMetadataStorage *objectstorage.ObjectStorage
	approversStorage           *objectstorage.ObjectStorage
	missingTransactionsStorage *objectstorage.ObjectStorage

	Events Events

	storeTransactionsWorkerPool async.WorkerPool
	solidifierWorkerPool        async.WorkerPool
	cleanupWorkerPool           async.WorkerPool
}

// Constructor for the tangle.
func New(storageId []byte) (result *Tangle) {
	result = &Tangle{
		storageId:                  storageId,
		transactionStorage:         objectstorage.New(append(storageId, storageprefix.TangleTransaction...), transaction.FromStorage),
		transactionMetadataStorage: objectstorage.New(append(storageId, storageprefix.TangleTransactionMetadata...), transactionmetadata.FromStorage),
		approversStorage:           objectstorage.New(append(storageId, storageprefix.TangleApprovers...), approvers.FromStorage),
		missingTransactionsStorage: objectstorage.New(append(storageId, storageprefix.TangleMissingTransaction...), missingtransaction.FromStorage),

		Events: *newEvents(),
	}

	result.solidifierWorkerPool.Tune(1024)

	return
}

func (tangle *Tangle) LoadSnapshot(snapshot *Snapshot) {
	fakeTransactionId := func(tx *transaction.Transaction, id transaction.Id) *transaction.Transaction {
		fakedTransaction := transaction.FromStorage(id[:])
		if err := fakedTransaction.UnmarshalBinary(tx.GetBytes()); err != nil {
			panic(err)
		}

		return fakedTransaction.(*transaction.Transaction)
	}

	for transactionId, addresses := range snapshot.SolidEntryPoints {
		if addresses == nil {
			tangle.AttachTransaction(fakeTransactionId(transaction.New(transaction.EmptyId, transaction.EmptyId, nil, data.New(nil)), transactionId))
		} else {
			valueTransfer := valuetransfer.New()

			for address, coloredBalance := range addresses {
				valueTransfer.AddOutput(address, coloredBalance)
			}

			tangle.AttachTransaction(fakeTransactionId(transaction.New(transaction.EmptyId, transaction.EmptyId, nil, valueTransfer), transactionId))
		}
	}
}

// Returns the storage id of this tangle (can be used to create ontologies that follow the storage of the main tangle).
func (tangle *Tangle) GetStorageId() []byte {
	return tangle.storageId
}

// Attaches a new transaction to the tangle.
func (tangle *Tangle) AttachTransaction(transaction *transaction.Transaction) {
	tangle.storeTransactionsWorkerPool.Submit(func() { tangle.storeTransactionWorker(transaction) })
}

// Retrieves a transaction from the tangle.
func (tangle *Tangle) GetTransaction(transactionId transaction.Id) *transaction.CachedTransaction {
	return &transaction.CachedTransaction{CachedObject: tangle.transactionStorage.Load(transactionId[:])}
}

// Retrieves the metadata of a transaction from the tangle.
func (tangle *Tangle) GetTransactionMetadata(transactionId transaction.Id) *transactionmetadata.CachedTransactionMetadata {
	return &transactionmetadata.CachedTransactionMetadata{CachedObject: tangle.transactionMetadataStorage.Load(transactionId[:])}
}

// Retrieves the approvers of a transaction from the tangle.
func (tangle *Tangle) GetApprovers(transactionId transaction.Id) *approvers.CachedApprovers {
	return &approvers.CachedApprovers{CachedObject: tangle.approversStorage.Load(transactionId[:])}
}

// Deletes a transaction from the tangle (i.e. for local snapshots)
func (tangle *Tangle) DeleteTransaction(transactionId transaction.Id) {
	tangle.GetTransaction(transactionId).Consume(func(object objectstorage.StorableObject) {
		currentTransaction := object.(*transaction.Transaction)

		tangle.GetApprovers(currentTransaction.GetTrunkTransactionId()).Consume(func(object objectstorage.StorableObject) {
			if _tmp := object.(*approvers.Approvers); _tmp.Remove(transactionId) && _tmp.Size() == 0 {
				_tmp.Delete()
			}
		})

		tangle.GetApprovers(currentTransaction.GetTrunkTransactionId()).Consume(func(object objectstorage.StorableObject) {
			if _tmp := object.(*approvers.Approvers); _tmp.Remove(transactionId) && _tmp.Size() == 0 {
				_tmp.Delete()
			}
		})
	})

	tangle.transactionStorage.Delete(transactionId[:])
	tangle.transactionMetadataStorage.Delete(transactionId[:])
	tangle.missingTransactionsStorage.Delete(transactionId[:])

	tangle.Events.TransactionRemoved.Trigger(transactionId)
}

// Marks the tangle as stopped, so it will not accept any new transactions (waits for all backgroundTasks to finish.
func (tangle *Tangle) Shutdown() *Tangle {
	tangle.storeTransactionsWorkerPool.ShutdownGracefully()
	tangle.solidifierWorkerPool.ShutdownGracefully()
	tangle.cleanupWorkerPool.ShutdownGracefully()

	return tangle
}

// Resets the database and deletes all objects (good for testing or "node resets").
func (tangle *Tangle) Prune() error {
	for _, storage := range []*objectstorage.ObjectStorage{
		tangle.transactionStorage,
		tangle.transactionMetadataStorage,
		tangle.approversStorage,
		tangle.missingTransactionsStorage,
	} {
		if err := storage.Prune(); err != nil {
			return err
		}
	}

	return nil
}

// Worker that stores the transactions and calls the corresponding "Storage events"
func (tangle *Tangle) storeTransactionWorker(tx *transaction.Transaction) {
	addTransactionToApprovers := func(transactionId transaction.Id, approvedTransactionId transaction.Id) {
		cachedApprovers := tangle.approversStorage.ComputeIfAbsent(approvedTransactionId[:], func([]byte) objectstorage.StorableObject {
			result := approvers.New(approvedTransactionId)

			result.SetModified()

			return result
		})

		if _tmp := cachedApprovers.Get(); _tmp != nil {
			if approversObject := _tmp.(*approvers.Approvers); approversObject != nil {
				approversObject.Add(transactionId)

				// if the approvers got "cleaned up" while being in cache, we make sure the object gets persisted again
				approversObject.Persist()
			}
		}

		cachedApprovers.Release()
	}

	var cachedTransaction *transaction.CachedTransaction
	if _tmp, transactionIsNew := tangle.transactionStorage.StoreIfAbsent(tx.GetStorageKey(), tx); !transactionIsNew {
		return
	} else {
		cachedTransaction = &transaction.CachedTransaction{CachedObject: _tmp}
	}

	transactionId := tx.GetId()

	cachedTransactionMetadata := &transactionmetadata.CachedTransactionMetadata{CachedObject: tangle.transactionMetadataStorage.Store(transactionmetadata.New(transactionId))}
	addTransactionToApprovers(transactionId, tx.GetTrunkTransactionId())
	addTransactionToApprovers(transactionId, tx.GetBranchTransactionId())

	if tangle.missingTransactionsStorage.DeleteIfPresent(transactionId[:]) {
		tangle.Events.MissingTransactionReceived.Trigger(transactionId)
	}

	tangle.Events.TransactionAttached.Trigger(cachedTransaction, cachedTransactionMetadata)

	tangle.solidifierWorkerPool.Submit(func() {
		tangle.solidifyTransactionWorker(cachedTransaction, cachedTransactionMetadata)
	})
}

// Worker that solidifies the transactions (recursively from past to present).
func (tangle *Tangle) solidifyTransactionWorker(cachedTransaction *transaction.CachedTransaction, cachedTransactionMetadata *transactionmetadata.CachedTransactionMetadata) {
	isTransactionMarkedAsSolid := func(transactionId transaction.Id) bool {
		if transactionId == transaction.EmptyId {
			return true
		}

		transactionMetadataCached := tangle.GetTransactionMetadata(transactionId)
		if transactionMetadata := transactionMetadataCached.Unwrap(); transactionMetadata == nil {
			transactionMetadataCached.Release()

			// if transaction is missing and was not reported as missing, yet
			if cachedMissingTransaction, missingTransactionStored := tangle.missingTransactionsStorage.StoreIfAbsent(transactionId[:], missingtransaction.New(transactionId)); missingTransactionStored {
				cachedMissingTransaction.Consume(func(object objectstorage.StorableObject) {
					tangle.monitorMissingTransactionWorker(object.(*missingtransaction.MissingTransaction).GetTransactionId())
				})
			}

			return false
		} else if !transactionMetadata.IsSolid() {
			transactionMetadataCached.Release()

			return false
		}
		transactionMetadataCached.Release()

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
		isTrunkSolid := isTransactionMarkedAsSolid(transaction.GetTrunkTransactionId())
		isBranchSolid := isTransactionMarkedAsSolid(transaction.GetBranchTransactionId())
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
	solidificationStack.PushBack([2]interface{}{cachedTransaction, cachedTransactionMetadata})

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
			tangle.Events.TransactionSolid.Trigger(currentCachedTransaction, currentCachedTransactionMetadata)

			tangle.GetApprovers(currentTransaction.GetId()).Consume(func(object objectstorage.StorableObject) {
				for approverTransactionId := range object.(*approvers.Approvers).Get() {
					solidificationStack.PushBack([2]interface{}{
						tangle.GetTransaction(approverTransactionId),
						tangle.GetTransactionMetadata(approverTransactionId),
					})
				}
			})
		}

		// release cached results
		currentCachedTransaction.Release()
		currentCachedTransactionMetadata.Release()
	}
}

// Worker that Monitors the missing transactions (by scheduling regular checks).
func (tangle *Tangle) monitorMissingTransactionWorker(transactionId transaction.Id) {
	var scheduleNextMissingCheck func(transactionId transaction.Id)
	scheduleNextMissingCheck = func(transactionId transaction.Id) {
		time.AfterFunc(MISSING_CHECK_INTERVAL, func() {
			tangle.missingTransactionsStorage.Load(transactionId[:]).Consume(func(object objectstorage.StorableObject) {
				missingTransaction := object.(*missingtransaction.MissingTransaction)

				if time.Since(missingTransaction.GetMissingSince()) >= MAX_MISSING_TIME_BEFORE_CLEANUP {
					tangle.cleanupWorkerPool.Submit(func() { tangle.cleanupWorker(missingTransaction.GetTransactionId()) })
				} else {
					tangle.Events.TransactionMissing.Trigger(transactionId)

					scheduleNextMissingCheck(transactionId)
				}
			})
		})
	}
	tangle.Events.TransactionMissing.Trigger(transactionId)

	scheduleNextMissingCheck(transactionId)
}

// Worker that recursively cleans up the approvers of a unsolidifiable missing transaction.
func (tangle *Tangle) cleanupWorker(transactionId transaction.Id) {
	cleanupStack := list.New()
	cleanupStack.PushBack(transactionId)

	for cleanupStack.Len() >= 1 {
		currentStackEntry := cleanupStack.Front()
		currentTransactionId := currentStackEntry.Value.(transaction.Id)
		cleanupStack.Remove(currentStackEntry)

		tangle.GetApprovers(currentTransactionId).Consume(func(object objectstorage.StorableObject) {
			for approverTransactionId := range object.(*approvers.Approvers).Get() {
				tangle.DeleteTransaction(currentTransactionId)

				cleanupStack.PushBack(approverTransactionId)
			}
		})
	}
}
