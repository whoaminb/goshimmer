package approvers

import (
	"encoding/binary"
	"sync"

	"github.com/iotaledger/goshimmer/packages/binary/tangle/model/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/types"

	"github.com/iotaledger/hive.go/objectstorage"
)

type Approvers struct {
	objectstorage.StorableObjectFlags

	transactionId  transaction.Id
	approvers      map[transaction.Id]types.Empty
	approversMutex sync.RWMutex
}

func New(transactionId transaction.Id) *Approvers {
	return &Approvers{
		transactionId: transactionId,
		approvers:     make(map[transaction.Id]types.Empty),
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

func (approvers *Approvers) Get() (result map[transaction.Id]types.Empty) {
	approvers.approversMutex.RLock()
	result = make(map[transaction.Id]types.Empty, len(approvers.approvers))
	for approverId := range approvers.approvers {
		result[approverId] = types.Void
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
			approvers.approvers[transactionId] = types.Void

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
	approvers.approversMutex.RLock()

	approversCount := len(approvers.approvers)
	result = make([]byte, 4+approversCount*transaction.IdLength)
	offset := 0

	binary.LittleEndian.PutUint32(result[offset:], uint32(approversCount))
	offset += 4

	for approverId := range approvers.approvers {
		marshaledBytes, marshalErr := approverId.MarshalBinary()
		if marshalErr != nil {
			err = marshalErr

			approvers.approversMutex.RUnlock()

			return
		}

		copy(result[offset:], marshaledBytes)
		offset += len(marshaledBytes)
	}

	approvers.approversMutex.RUnlock()

	return
}

func (approvers *Approvers) UnmarshalBinary(data []byte) (err error) {
	approvers.approvers = make(map[transaction.Id]types.Empty)
	offset := 0

	approversCount := int(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4

	for i := 0; i < approversCount; i++ {
		var approverId transaction.Id
		if err = approverId.UnmarshalBinary(data[offset:]); err != nil {
			return
		}
		offset += transaction.IdLength

		approvers.approvers[approverId] = types.Void
	}

	return
}
