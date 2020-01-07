package model

import (
	"encoding/binary"
	"sync"

	"github.com/iotaledger/goshimmer/packages/binary/types"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
	"github.com/iotaledger/hive.go/objectstorage"
)

type Consumers struct {
	objectstorage.StorableObjectFlags

	transferId     transfer.Id
	consumers      map[transfer.Id]types.Empty
	consumersMutex sync.RWMutex
}

func NewConsumers(transferId transfer.Id) *Consumers {
	return &Consumers{
		transferId: transferId,
		consumers:  make(map[transfer.Id]types.Empty),
	}
}

// Get's called when we restore the approvers from storage. The bytes and the content will be unmarshaled by an external
// caller (the objectStorage factory).
func FromStorage(id []byte) (result objectstorage.StorableObject) {
	var transferId transfer.Id
	copy(transferId[:], id)

	result = &Consumers{
		transferId: transferId,
	}

	return
}

func (consumers *Consumers) GetTransferId() transfer.Id {
	return consumers.transferId
}

func (consumers *Consumers) Get() (result map[transfer.Id]types.Empty) {
	consumers.consumersMutex.RLock()
	result = make(map[transfer.Id]types.Empty, len(consumers.consumers))
	for approverId := range consumers.consumers {
		result[approverId] = types.Void
	}
	consumers.consumersMutex.RUnlock()

	return
}

func (consumers *Consumers) Add(transferId transfer.Id) (modified bool) {
	consumers.consumersMutex.RLock()
	if _, exists := consumers.consumers[transferId]; !exists {
		consumers.consumersMutex.RUnlock()

		consumers.consumersMutex.Lock()
		if _, exists := consumers.consumers[transferId]; !exists {
			consumers.consumers[transferId] = types.Void

			modified = true

			consumers.SetModified()
		}
		consumers.consumersMutex.Unlock()
	} else {
		consumers.consumersMutex.RUnlock()
	}

	return
}

func (consumers *Consumers) Remove(transferId transfer.Id) (modified bool) {
	consumers.consumersMutex.RLock()
	if _, exists := consumers.consumers[transferId]; exists {
		consumers.consumersMutex.RUnlock()

		consumers.consumersMutex.Lock()
		if _, exists := consumers.consumers[transferId]; exists {
			delete(consumers.consumers, transferId)

			modified = true

			consumers.SetModified()
		}
		consumers.consumersMutex.Unlock()
	} else {
		consumers.consumersMutex.RUnlock()
	}

	return
}

func (consumers *Consumers) Size() (result int) {
	consumers.consumersMutex.RLock()
	result = len(consumers.consumers)
	consumers.consumersMutex.RUnlock()

	return
}

func (consumers *Consumers) GetStorageKey() []byte {
	transferId := consumers.GetTransferId()

	return transferId[:]
}

func (consumers *Consumers) Update(other objectstorage.StorableObject) {
	panic("approvers should never be overwritten and only stored once to optimize IO")
}

func (consumers *Consumers) MarshalBinary() (result []byte, err error) {
	consumers.consumersMutex.RLock()

	approversCount := len(consumers.consumers)
	result = make([]byte, 4+approversCount*transfer.IdLength)
	offset := 0

	binary.LittleEndian.PutUint32(result[offset:], uint32(approversCount))
	offset += 4

	for approverId := range consumers.consumers {
		marshaledBytes, marshalErr := approverId.MarshalBinary()
		if marshalErr != nil {
			err = marshalErr

			consumers.consumersMutex.RUnlock()

			return
		}

		copy(result[offset:], marshaledBytes)
		offset += len(marshaledBytes)
	}

	consumers.consumersMutex.RUnlock()

	return
}

func (consumers *Consumers) UnmarshalBinary(data []byte) (err error) {
	consumers.consumers = make(map[transfer.Id]types.Empty)
	offset := 0

	approversCount := int(binary.LittleEndian.Uint32(data[offset:]))
	offset += 4

	for i := 0; i < approversCount; i++ {
		var approverId transfer.Id
		if err = approverId.UnmarshalBinary(data[offset:]); err != nil {
			return
		}
		offset += transfer.IdLength

		consumers.consumers[approverId] = types.Void
	}

	return
}
