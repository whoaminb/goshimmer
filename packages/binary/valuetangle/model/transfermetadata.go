package model

import (
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"

	"github.com/iotaledger/hive.go/objectstorage"
)

type TransferMetadata struct {
	objectstorage.StorableObjectFlags

	transferId         transfer.Id
	receivedTime       time.Time
	solid              bool
	solidificationTime time.Time

	solidMutex sync.RWMutex
}

func NewTransferMetadata(transferId transfer.Id) *TransferMetadata {
	return &TransferMetadata{
		transferId:   transferId,
		receivedTime: time.Now(),
	}
}

func TransferMetadataFromStorage(id []byte) objectstorage.StorableObject {
	result := &TransferMetadata{}
	copy(result.transferId[:], id)

	return result
}

func (transferMetadata *TransferMetadata) IsSolid() (result bool) {
	transferMetadata.solidMutex.RLock()
	result = transferMetadata.solid
	transferMetadata.solidMutex.RUnlock()

	return
}

func (transferMetadata *TransferMetadata) SetSolid(solid bool) (modified bool) {
	transferMetadata.solidMutex.RLock()
	if transferMetadata.solid != solid {
		transferMetadata.solidMutex.RUnlock()

		transferMetadata.solidMutex.Lock()
		if transferMetadata.solid != solid {
			transferMetadata.solid = solid

			transferMetadata.SetModified()

			modified = true
		}
		transferMetadata.solidMutex.Unlock()

	} else {
		transferMetadata.solidMutex.RUnlock()
	}

	return
}

func (transferMetadata *TransferMetadata) GetStorageKey() []byte {
	return transferMetadata.transferId[:]
}

func (transferMetadata *TransferMetadata) Update(other objectstorage.StorableObject) {
	panic("TransferMetadata should never be overwritten")
}

func (transferMetadata *TransferMetadata) MarshalBinary() ([]byte, error) {
	return nil, nil
}

func (transferMetadata *TransferMetadata) UnmarshalBinary([]byte) error {
	return nil
}
