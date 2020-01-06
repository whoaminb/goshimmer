package missingtransaction

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/binary/transaction"

	"github.com/iotaledger/hive.go/objectstorage"
)

type MissingTransaction struct {
	objectstorage.StorableObjectFlags
	storageKey []byte

	id           transaction.Id
	missingSince time.Time
}

func New(id transaction.Id) *MissingTransaction {
	return &MissingTransaction{
		storageKey:   id[:],
		id:           id,
		missingSince: time.Now(),
	}
}

func FromStorage(key []byte) objectstorage.StorableObject {
	result := &MissingTransaction{
		storageKey: make([]byte, len(key)),
	}
	copy(result.storageKey, key)

	return result
}

func (missingTransaction *MissingTransaction) GetId() transaction.Id {
	return missingTransaction.id
}

func (missingTransaction *MissingTransaction) GetMissingSince() time.Time {
	return missingTransaction.missingSince
}

func (missingTransaction *MissingTransaction) GetStorageKey() []byte {
	return missingTransaction.storageKey
}

func (missingTransaction *MissingTransaction) Update(other objectstorage.StorableObject) {
	panic("missing transactions should never be overwritten and only stored once to optimize IO")
}

func (missingTransaction *MissingTransaction) MarshalBinary() (result []byte, err error) {
	return missingTransaction.missingSince.MarshalBinary()
}

func (missingTransaction *MissingTransaction) UnmarshalBinary(data []byte) (err error) {
	copy(missingTransaction.id[:], missingTransaction.storageKey)

	return missingTransaction.missingSince.UnmarshalBinary(data)
}
