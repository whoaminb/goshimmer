package transactionmetadata

import (
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/hive.go/objectstorage"
)

type TransactionMetadata struct {
	objectstorage.StorableObjectFlags

	transactionId      transaction.Id
	receivedTime       time.Time
	solid              bool
	solidificationTime time.Time

	solidMutex sync.RWMutex
}

func New(transactionId transaction.Id) *TransactionMetadata {
	return &TransactionMetadata{
		transactionId: transactionId,
		receivedTime:  time.Now(),
	}
}

func (transactionMetadata *TransactionMetadata) IsSolid() (result bool) {
	transactionMetadata.solidMutex.RLock()
	result = transactionMetadata.solid
	transactionMetadata.solidMutex.RUnlock()

	return
}

func (transactionMetadata *TransactionMetadata) SetSolid(solid bool) (modified bool) {
	transactionMetadata.solidMutex.RLock()
	if transactionMetadata.solid != solid {
		transactionMetadata.solidMutex.RUnlock()

		transactionMetadata.solidMutex.Lock()
		if transactionMetadata.solid != solid {
			transactionMetadata.solid = solid

			transactionMetadata.SetModified()

			modified = true
		}
		transactionMetadata.solidMutex.Unlock()

	} else {
		transactionMetadata.solidMutex.RUnlock()
	}

	return
}

func (transactionMetadata *TransactionMetadata) GetStorageKey() []byte {
	return transactionMetadata.transactionId[:]
}

func (transactionMetadata *TransactionMetadata) Update(other objectstorage.StorableObject) {

}

func (transactionMetadata *TransactionMetadata) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (transactionMetadata *TransactionMetadata) UnmarshalBinary([]byte) error {
	return nil
}
