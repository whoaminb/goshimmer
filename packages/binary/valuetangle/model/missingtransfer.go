package model

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"

	"github.com/iotaledger/hive.go/objectstorage"
)

type MissingTransfer struct {
	objectstorage.StorableObjectFlags

	transferId   transfer.Id
	missingSince time.Time
}

func NewMissingTransfer(transferId transfer.Id) *MissingTransfer {
	return &MissingTransfer{
		transferId:   transferId,
		missingSince: time.Now(),
	}
}

func MissingTransferFromStorage(key []byte) objectstorage.StorableObject {
	result := &MissingTransfer{}
	copy(result.transferId[:], key)

	return result
}

func (missingTransfer *MissingTransfer) GetTransferId() transfer.Id {
	return missingTransfer.transferId
}

func (missingTransfer *MissingTransfer) GetMissingSince() time.Time {
	return missingTransfer.missingSince
}

func (missingTransfer *MissingTransfer) GetStorageKey() []byte {
	return missingTransfer.transferId[:]
}

func (missingTransfer *MissingTransfer) Update(other objectstorage.StorableObject) {
	panic("missing transfer should never be overwritten and only stored once to optimize IO")
}

func (missingTransfer *MissingTransfer) MarshalBinary() (result []byte, err error) {
	return missingTransfer.missingSince.MarshalBinary()
}

func (missingTransfer *MissingTransfer) UnmarshalBinary(data []byte) (err error) {
	return missingTransfer.missingSince.UnmarshalBinary(data)
}
