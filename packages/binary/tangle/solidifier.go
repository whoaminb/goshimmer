package tangle

import (
	"container/list"

	"github.com/iotaledger/goshimmer/packages/binary/tangle/approvers"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/transactionmetadata"
	"github.com/iotaledger/hive.go/async"
	"github.com/iotaledger/hive.go/objectstorage"
)

type solidifier struct {
	tangle *Tangle

	workerPool async.WorkerPool
}

func newSolidifier(tangle *Tangle) (result *solidifier) {
	result = &solidifier{
		tangle: tangle,
	}

	result.workerPool.Tune(1024)

	return
}

func (solidifier *solidifier) Shutdown() {
	solidifier.workerPool.ShutdownGracefully()
}

func (solidifier *solidifier) Solidify(cachedTransaction *objectstorage.CachedObject, cachedTransactionMetadata *objectstorage.CachedObject) {
	solidifier.workerPool.Submit(func() { solidifier.solidify(cachedTransaction, cachedTransactionMetadata) })
}

func (solidifier *solidifier) solidify(cachedTransaction *objectstorage.CachedObject, cachedTransactionMetadata *objectstorage.CachedObject) {
	// initialize the stack
	solidificationStack := list.New()
	solidificationStack.PushBack([2]*objectstorage.CachedObject{cachedTransaction, cachedTransactionMetadata})

	// process transactions that are supposed to be checked for solidity recursively
	for solidificationStack.Len() > 0 {
		// pop first element from stack
		currentSolidificationEntry := solidificationStack.Front()
		currentCachedTransaction := currentSolidificationEntry.Value.([2]*objectstorage.CachedObject)[0]
		currentCachedTransactionMetadata := currentSolidificationEntry.Value.([2]*objectstorage.CachedObject)[1]
		solidificationStack.Remove(currentSolidificationEntry)

		// retrieve transaction from cached result
		var currentTransaction *transaction.Transaction
		if _tmp := currentCachedTransaction.Get(); _tmp != nil {
			currentTransaction = _tmp.(*transaction.Transaction)
		} else {
			currentCachedTransaction.Release()
			currentCachedTransactionMetadata.Release()

			continue
		}

		// retrieve metadata from cached result
		var currentTransactionMetadata *transactionmetadata.TransactionMetadata
		if _tmp := currentCachedTransactionMetadata.Get(); _tmp != nil {
			currentTransactionMetadata = _tmp.(*transactionmetadata.TransactionMetadata)
		} else {
			currentCachedTransaction.Release()
			currentCachedTransactionMetadata.Release()

			continue
		}

		// if current transaction is solid and was not marked as solid before: mark as solid and propagate
		if solidifier.isTransactionSolid(currentTransaction, currentTransactionMetadata) && currentTransactionMetadata.SetSolid(true) {
			solidifier.tangle.Events.TransactionSolid.Trigger(currentCachedTransaction, currentCachedTransactionMetadata)

			solidifier.tangle.GetApprovers(currentTransaction.GetId()).Consume(func(object objectstorage.StorableObject) {
				for approverTransactionId := range object.(*approvers.Approvers).Get() {
					solidificationStack.PushBack([2]*objectstorage.CachedObject{
						solidifier.tangle.GetTransaction(approverTransactionId),
						solidifier.tangle.GetTransactionMetadata(approverTransactionId),
					})
				}
			})
		}

		// release cached results
		currentCachedTransaction.Release()
		currentCachedTransactionMetadata.Release()
	}
}

func (solidifier *solidifier) isTransactionSolid(transaction *transaction.Transaction, transactionMetadata *transactionmetadata.TransactionMetadata) bool {
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
	if solidifier.isTransactionSolidInTangle(transaction.GetTrunkTransactionId()) && solidifier.isTransactionSolidInTangle(transaction.GetBranchTransactionId()) {
		// 2. check payload solidity
		return true
	}

	return false
}

func (solidifier *solidifier) isTransactionSolidInTangle(transactionId transaction.Id) bool {
	if transactionId != transaction.EmptyId {
		cachedTransactionMetadata := solidifier.tangle.GetTransactionMetadata(transactionId)

		if transactionMetadata := cachedTransactionMetadata.Get().(*transactionmetadata.TransactionMetadata); transactionMetadata == nil || transactionMetadata.IsDeleted() || !transactionMetadata.IsSolid() {
			cachedTransactionMetadata.Release()

			return false
		}

		cachedTransactionMetadata.Release()
	}

	return true
}
