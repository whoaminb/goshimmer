package tangle

import (
	"github.com/iotaledger/goshimmer/packages/binary/tangle/approvers"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/transactionmetadata"
	"github.com/iotaledger/goshimmer/packages/storageprefix"
	"github.com/iotaledger/goshimmer/packages/stringify"
	"github.com/iotaledger/hive.go/async"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/objectstorage"
	"github.com/pkg/errors"
)

type Tangle struct {
	solidifier                 *solidifier
	transactionStorage         *objectstorage.ObjectStorage
	transactionMetadataStorage *objectstorage.ObjectStorage
	approversStorage           *objectstorage.ObjectStorage

	Events tangleEvents

	verifyTransactionsWorkerPool async.WorkerPool
	storeTransactionsWorkerPool  async.WorkerPool
}

func New(storageId []byte) (result *Tangle) {
	result = &Tangle{
		transactionStorage:         objectstorage.New(append(storageId, storageprefix.TangleTransaction...), transactionFactory),
		transactionMetadataStorage: objectstorage.New(append(storageId, storageprefix.TangleTransactionMetadata...), transactionFactory),
		approversStorage:           objectstorage.New(append(storageId, storageprefix.TangleApprovers...), approversFactory),

		Events: tangleEvents{
			TransactionAttached: events.NewEvent(func(handler interface{}, params ...interface{}) {
				cachedTransaction := params[0].(*objectstorage.CachedObject)
				cachedTransactionMetadata := params[1].(*objectstorage.CachedObject)

				cachedTransaction.RegisterConsumer()
				cachedTransactionMetadata.RegisterConsumer()

				handler.(func(*objectstorage.CachedObject, *objectstorage.CachedObject))(cachedTransaction, cachedTransactionMetadata)
			}),
			TransactionSolid: events.NewEvent(func(handler interface{}, params ...interface{}) {
				cachedTransaction := params[0].(*objectstorage.CachedObject)
				cachedTransactionMetadata := params[1].(*objectstorage.CachedObject)

				cachedTransaction.RegisterConsumer()
				cachedTransactionMetadata.RegisterConsumer()

				handler.(func(*objectstorage.CachedObject, *objectstorage.CachedObject))(cachedTransaction, cachedTransactionMetadata)
			}),
			Error: events.NewEvent(func(handler interface{}, params ...interface{}) {
				handler.(func(error))(params[0].(error))
			}),
		},
	}

	result.solidifier = newSolidifier(result)

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
	tangle.verifyTransactionsWorkerPool.Submit(func() { tangle.verifyTransaction(transaction) })
}

func (tangle *Tangle) GetTransaction(transactionId transaction.Id) *objectstorage.CachedObject {
	return tangle.transactionStorage.Load(transactionId[:])
}

func (tangle *Tangle) GetTransactionMetadata(transactionId transaction.Id) *objectstorage.CachedObject {
	return tangle.transactionMetadataStorage.Load(transactionId[:])
}

func (tangle *Tangle) GetApprovers(transactionId transaction.Id) *objectstorage.CachedObject {
	return tangle.approversStorage.Load(transactionId[:])
}

func (tangle *Tangle) verifyTransaction(transaction *transaction.Transaction) {
	if !transaction.VerifySignature() {
		tangle.Events.Error.Trigger(errors.New("transaction with id " + stringify.Interface(transaction.GetId()) + " has an invalid signature"))

		return
	}

	tangle.storeTransactionsWorkerPool.Submit(func() { tangle.storeTransaction(transaction) })
}

func (tangle *Tangle) storeTransaction(transaction *transaction.Transaction) {
	cachedTransaction, transactionIsNew := tangle.transactionStorage.StoreIfAbsent(transaction.GetStorageKey(), transaction)
	if !transactionIsNew {
		return
	}

	cachedTransactionMetadata := tangle.createTransactionMetadata(transaction)

	tangle.addTransactionToApprovers(transaction, transaction.GetTrunkTransactionId())
	tangle.addTransactionToApprovers(transaction, transaction.GetBranchTransactionId())

	tangle.solidifier.Solidify(cachedTransaction, cachedTransactionMetadata)
}

// Marks the tangle as stopped, so it will not accept any new transactions, and then waits for all backgroundTasks to
// finish.
func (tangle *Tangle) Shutdown() *Tangle {
	tangle.verifyTransactionsWorkerPool.ShutdownGracefully()
	tangle.storeTransactionsWorkerPool.ShutdownGracefully()

	tangle.solidifier.Shutdown()

	return tangle
}

func (tangle *Tangle) createTransactionMetadata(transaction *transaction.Transaction) *objectstorage.CachedObject {
	transactionMetadata := transactionmetadata.New(transaction.GetId())

	return tangle.transactionMetadataStorage.Store(transactionMetadata)
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
