package tangle

import (
	"github.com/iotaledger/goshimmer/packages/binary/tangle/approvers"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/hive.go/objectstorage"
)

type Tangle struct {
	transactionStorage *objectstorage.ObjectStorage
	approversStorage   *objectstorage.ObjectStorage
}

func New(storageId string) *Tangle {
	return &Tangle{
		transactionStorage: objectstorage.New(storageId+"TANGLE_TRANSACTION_STORAGE", transactionFactory),
		approversStorage:   objectstorage.New(storageId+"TANGLE_APPROVERS_STORAGE", approversFactory),
	}
}

func transactionFactory(key []byte) objectstorage.StorableObject {
	result := transaction.FromStorage(key)

	return result
}

func approversFactory(key []byte) objectstorage.StorableObject {
	result := approvers.FromStorage(key)

	return result
}
