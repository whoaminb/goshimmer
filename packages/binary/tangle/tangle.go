package tangle

import (
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/hive.go/objectstorage"
)

type Tangle struct {
	transactionStorage *objectstorage.ObjectStorage
}

func New(storageId string) *Tangle {
	return &Tangle{
		transactionStorage: objectstorage.New(storageId+"TANGLE_TRANSACTION_STORAGE", transactionFactory),
	}
}

func transactionFactory(key []byte) objectstorage.StorableObject {
	result := transaction.FromStorage(key)

	return result
}
