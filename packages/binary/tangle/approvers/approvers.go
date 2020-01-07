package approvers

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/hive.go/objectstorage"
)

type Approvers struct {
	objectstorage.StorableObjectFlags

	transactionId  transaction.Id
	approvers      map[transaction.Id]empty
	approversMutex sync.RWMutex
}

func New(transactionId transaction.Id) *Approvers {
	return &Approvers{
		transactionId: transactionId,
		approvers:     make(map[transaction.Id]empty),
	}
}

// Get's called when we restore the approvers from storage. The bytes and the content will be unmarshaled by an external
// caller (the objectStorage factory).
func FromStorage(id []byte) (result objectstorage.StorableObject) {
	var transactionId transaction.Id
	copy(transactionId[:], id)

	result = &Approvers{
		transactionId: transactionId,
	}

	return
}

func (approvers *Approvers) GetTransactionId() transaction.Id {
	return approvers.transactionId
}

func (approvers *Approvers) Get() (result map[transaction.Id]empty) {
	approvers.approversMutex.RLock()
	result = make(map[transaction.Id]empty, len(approvers.approvers))
	for approverId := range approvers.approvers {
		result[approverId] = void
	}
	approvers.approversMutex.RUnlock()

	return
}

func (approvers *Approvers) Add(transactionId transaction.Id) (modified bool) {
	approvers.approversMutex.RLock()
	if _, exists := approvers.approvers[transactionId]; !exists {
		approvers.approversMutex.RUnlock()

		approvers.approversMutex.Lock()
		if _, exists := approvers.approvers[transactionId]; !exists {
			approvers.approvers[transactionId] = void

			modified = true

			approvers.SetModified()
		}
		approvers.approversMutex.Unlock()
	} else {
		approvers.approversMutex.RUnlock()
	}

	return
}

func (approvers *Approvers) Remove(transactionId transaction.Id) (modified bool) {
	approvers.approversMutex.RLock()
	if _, exists := approvers.approvers[transactionId]; exists {
		approvers.approversMutex.RUnlock()

		approvers.approversMutex.Lock()
		if _, exists := approvers.approvers[transactionId]; exists {
			delete(approvers.approvers, transactionId)

			modified = true

			approvers.SetModified()
		}
		approvers.approversMutex.Unlock()
	} else {
		approvers.approversMutex.RUnlock()
	}

	return
}

func (approvers *Approvers) Size() (result int) {
	approvers.approversMutex.RLock()
	result = len(approvers.approvers)
	approvers.approversMutex.RUnlock()

	return
}

func (approvers *Approvers) GetStorageKey() []byte {
	transactionId := approvers.GetTransactionId()

	return transactionId[:]
}

func (approvers *Approvers) Update(other objectstorage.StorableObject) {
	panic("approvers should never be overwritten and only stored once to optimize IO")
}

func (approvers *Approvers) MarshalBinary() (result []byte, err error) {
	return
}

func (approvers *Approvers) UnmarshalBinary(data []byte) (err error) {
	approvers.approvers = make(map[transaction.Id]empty)

	return
}
