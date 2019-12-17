package transaction

import (
	"sync"

	"github.com/iotaledger/hive.go/objectstorage"

	"golang.org/x/crypto/blake2b"
)

type Transaction struct {
	// base functionality of StorableObject
	objectstorage.StorableObjectFlags

	// core properties (they are part of the transaction when being sent)
	trunkTransactionId  Id
	branchTransactionId Id
	payload             Payload

	// derived properties
	id             *Id
	idMutex        sync.RWMutex
	payloadId      *PayloadId
	payloadIdMutex sync.RWMutex
	bytes          []byte
	bytesMutex     sync.RWMutex
}

func FromStorage(id []byte) (result *Transaction) {
	var transactionId Id
	copy(transactionId[:], id)

	result = &Transaction{
		id: &transactionId,
	}

	return
}

func (transaction *Transaction) GetId() (result Id) {
	transaction.idMutex.RLock()
	if transaction.id == nil {
		transaction.idMutex.RLock()

		transaction.idMutex.Lock()
		if transaction.id == nil {
			result = transaction.calculateTransactionId()

			transaction.id = &result
		} else {
			result = *transaction.id
		}
		transaction.idMutex.Unlock()
	} else {
		result = *transaction.id

		transaction.idMutex.RLock()
	}

	return
}

func (transaction *Transaction) GetPayloadId() (result PayloadId) {
	transaction.payloadIdMutex.RLock()
	if transaction.payloadId == nil {
		transaction.payloadIdMutex.RLock()

		transaction.payloadIdMutex.Lock()
		if transaction.payloadId == nil {
			result = transaction.calculatePayloadId()

			transaction.payloadId = &result
		} else {
			result = *transaction.payloadId
		}
		transaction.payloadIdMutex.Unlock()
	} else {
		result = *transaction.payloadId

		transaction.payloadIdMutex.RLock()
	}

	return
}

func (transaction *Transaction) GetBytes() (result []byte) {
	transaction.bytesMutex.RLock()
	if transaction.bytes == nil {
		transaction.bytesMutex.RLock()

		transaction.bytesMutex.Lock()
		if transaction.bytes == nil {
			var err error

			if result, err = transaction.MarshalBinary(); err != nil {
				// this should never happen
				panic(err)
			}

			transaction.bytes = result
		} else {
			result = transaction.bytes
		}
		transaction.bytesMutex.Unlock()
	} else {
		result = transaction.bytes

		transaction.bytesMutex.RLock()
	}

	return
}

func (transaction *Transaction) calculateTransactionId() Id {
	payloadId := transaction.GetPayloadId()

	hashBase := make([]byte, transactionIdLength+transactionIdLength+payloadIdLength)
	offset := 0

	copy(hashBase[offset:], transaction.trunkTransactionId[:])
	offset += transactionIdLength

	copy(hashBase[offset:], transaction.branchTransactionId[:])
	offset += transactionIdLength

	copy(hashBase[offset:], payloadId[:])
	offset += payloadIdLength

	return blake2b.Sum512(hashBase)
}

func (transaction *Transaction) calculatePayloadId() PayloadId {
	bytes := transaction.GetBytes()

	return blake2b.Sum512(bytes[2*transactionIdLength:])
}

func (transaction *Transaction) MarshalBinary() (result []byte, err error) {
	result = make([]byte, 2*transactionIdLength)

	if serializedPayload, serializationErr := transaction.payload.MarshalBinary(); serializationErr != nil {
		err = serializationErr

		return
	} else {
		result = append(result, serializedPayload...)
	}

	return
}

// TODO: FINISH
func (transaction *Transaction) UnmarshalBinary(date []byte) (err error) {
	return
}

func (transaction *Transaction) GetStorageKey() []byte {
	transactionId := transaction.GetId()

	return transactionId[:]
}

// TODO: FINISH
func (transaction *Transaction) Update(other objectstorage.StorableObject) {}
