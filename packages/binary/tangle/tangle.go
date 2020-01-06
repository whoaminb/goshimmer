package tangle

import (
	"container/list"
	"fmt"

	"github.com/iotaledger/goshimmer/packages/binary/tangle/approvers"
	"github.com/iotaledger/goshimmer/packages/binary/tangle/missingtransaction"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/transactionmetadata"
	"github.com/iotaledger/goshimmer/packages/storageprefix"
	"github.com/iotaledger/hive.go/async"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/objectstorage"
)

type Tangle struct {
	transactionStorage         *objectstorage.ObjectStorage
	transactionMetadataStorage *objectstorage.ObjectStorage
	approversStorage           *objectstorage.ObjectStorage
	missingTransactionsStorage *objectstorage.ObjectStorage

	Events tangleEvents

	storeTransactionsWorkerPool async.WorkerPool
	solidifierWorkerPool        async.WorkerPool
}

func New(storageId []byte) (result *Tangle) {
	result = &Tangle{
		transactionStorage:         objectstorage.New(append(storageId, storageprefix.TangleTransaction...), transactionFactory),
		transactionMetadataStorage: objectstorage.New(append(storageId, storageprefix.TangleTransactionMetadata...), transactionFactory),
		approversStorage:           objectstorage.New(append(storageId, storageprefix.TangleApprovers...), approversFactory),
		missingTransactionsStorage: objectstorage.New(append(storageId, storageprefix.TangleTransaction...), missingtransaction.FromStorage),

		Events: tangleEvents{
			TransactionAttached: events.NewEvent(func(handler interface{}, params ...interface{}) {
				cachedTransaction := params[0].(*transaction.CachedTransaction)
				cachedTransactionMetadata := params[1].(*transactionmetadata.CachedTransactionMetadata)

				cachedTransaction.RegisterConsumer()
				cachedTransactionMetadata.RegisterConsumer()

				handler.(func(*transaction.CachedTransaction, *transactionmetadata.CachedTransactionMetadata))(cachedTransaction, cachedTransactionMetadata)
			}),
			TransactionSolid: events.NewEvent(func(handler interface{}, params ...interface{}) {
				cachedTransaction := params[0].(*transaction.CachedTransaction)
				cachedTransactionMetadata := params[1].(*transactionmetadata.CachedTransactionMetadata)

				cachedTransaction.RegisterConsumer()
				cachedTransactionMetadata.RegisterConsumer()

				handler.(func(*transaction.CachedTransaction, *transactionmetadata.CachedTransactionMetadata))(cachedTransaction, cachedTransactionMetadata)
			}),
			Error: events.NewEvent(func(handler interface{}, params ...interface{}) {
				handler.(func(error))(params[0].(error))
			}),
		},
	}

	result.solidifierWorkerPool.Tune(1024)

	return
}

func (tangle *Tangle) Prune() error {
	if err := tangle.transactionStorage.Prune(); err != nil {
		return err
	}

	if err := tangle.transactionMetadataStorage.Prune(); err != nil {
		return err
	}

	if err := tangle.approversStorage.Prune(); err != nil {
		return err
	}

	return nil
}

func (tangle *Tangle) AttachTransaction(transaction *transaction.Transaction) {
	tangle.storeTransactionsWorkerPool.Submit(func() { tangle.storeTransaction(transaction) })
}

func (tangle *Tangle) GetTransaction(transactionId transaction.Id) *transaction.CachedTransaction {
	return &transaction.CachedTransaction{CachedObject: tangle.transactionStorage.Load(transactionId[:])}
}

func (tangle *Tangle) GetTransactionMetadata(transactionId transaction.Id) *transactionmetadata.CachedTransactionMetadata {
	return &transactionmetadata.CachedTransactionMetadata{CachedObject: tangle.transactionMetadataStorage.Load(transactionId[:])}
}

func (tangle *Tangle) GetApprovers(transactionId transaction.Id) *approvers.CachedApprovers {
	return &approvers.CachedApprovers{CachedObject: tangle.approversStorage.Load(transactionId[:])}
}

// Marks the tangle as stopped, so it will not accept any new transactions, and then waits for all backgroundTasks to
// finish.
func (tangle *Tangle) Shutdown() *Tangle {
	tangle.storeTransactionsWorkerPool.ShutdownGracefully()
	tangle.solidifierWorkerPool.ShutdownGracefully()

	return tangle
}

func (tangle *Tangle) storeTransaction(tx *transaction.Transaction) {
	cachedTransaction, transactionIsNew := tangle.transactionStorage.StoreIfAbsent(tx.GetStorageKey(), tx)
	if !transactionIsNew {
		return
	}

	cachedTransactionMetadata := tangle.createTransactionMetadata(tx)
	tangle.addTransactionToApprovers(tx, tx.GetTrunkTransactionId())
	tangle.addTransactionToApprovers(tx, tx.GetBranchTransactionId())

	transactionId := tx.GetId()
	if tangle.missingTransactionsStorage.DeleteIfPresent(transactionId[:]) {
		fmt.Println("MISSING TRANSACTION RECEIVED")
	}

	tangle.solidifierWorkerPool.Submit(func() {
		tangle.solidify(&transaction.CachedTransaction{CachedObject: cachedTransaction}, cachedTransactionMetadata)
	})
}

// Payloads can have different solidification rules and it might happen, that an external process needs to "manually
// trigger" the solidification checks of a transaction (to update it's solidification status).
func (tangle *Tangle) Solidify(transactionId transaction.Id) {
	tangle.solidifierWorkerPool.Submit(func() {
		tangle.solidify(tangle.GetTransaction(transactionId), tangle.GetTransactionMetadata(transactionId))
	})
}

func (tangle *Tangle) solidify(cachedTransaction *transaction.CachedTransaction, cachedTransactionMetadata *transactionmetadata.CachedTransactionMetadata) {
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
		if tangle.isTransactionSolid(currentTransaction, currentTransactionMetadata) && currentTransactionMetadata.SetSolid(true) {
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

func (tangle *Tangle) isTransactionSolid(transaction *transaction.Transaction, transactionMetadata *transactionmetadata.TransactionMetadata) bool {
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
	isTrunkSolid := tangle.isTransactionMarkedAsSolid(transaction.GetTrunkTransactionId())
	isBranchSolid := tangle.isTransactionMarkedAsSolid(transaction.GetBranchTransactionId())
	if isTrunkSolid && isBranchSolid {
		// 2. check payload solidity
		return true
	}

	return false
}

func (tangle *Tangle) isTransactionMarkedAsSolid(transactionId transaction.Id) bool {
	if transactionId == transaction.EmptyId {
		return true
	}

	cachedTransactionMetadata := tangle.GetTransactionMetadata(transactionId)
	if transactionMetadata := cachedTransactionMetadata.Unwrap(); transactionMetadata == nil {
		cachedTransactionMetadata.Release()

		if _, missingTransactionStored := tangle.missingTransactionsStorage.StoreIfAbsent(transactionId[:], &missingtransaction.MissingTransaction{}); missingTransactionStored {
			// Trigger
			fmt.Println("MISSING TX EVENT")
		}
		// transaction is missing -> add to solidifier

		return false
	} else if !transactionMetadata.IsSolid() {
		cachedTransactionMetadata.Release()

		return false
	}
	cachedTransactionMetadata.Release()

	return true
}

func (tangle *Tangle) createTransactionMetadata(transaction *transaction.Transaction) *transactionmetadata.CachedTransactionMetadata {
	transactionMetadata := transactionmetadata.New(transaction.GetId())

	return &transactionmetadata.CachedTransactionMetadata{CachedObject: tangle.transactionMetadataStorage.Store(transactionMetadata)}
}

func (tangle *Tangle) addTransactionToApprovers(transaction *transaction.Transaction, trunkTransactionId transaction.Id) {
	tangle.approversStorage.ComputeIfAbsent(trunkTransactionId[:], func([]byte) objectstorage.StorableObject {
		return approvers.New(trunkTransactionId)
	}).Consume(func(object objectstorage.StorableObject) {
		object.(*approvers.Approvers).Add(transaction.GetId())
	})
}

func transactionFactory(key []byte) objectstorage.StorableObject {
	return transaction.FromStorage(key)
}

func approversFactory(key []byte) objectstorage.StorableObject {
	return approvers.FromStorage(key)
}
