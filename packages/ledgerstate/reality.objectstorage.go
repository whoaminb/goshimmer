package ledgerstate

import (
	"encoding/binary"

	"github.com/iotaledger/hive.go/objectstorage"
)

func (reality *Reality) GetStorageKey() []byte {
	return reality.storageKey
}

func (reality *Reality) Update(other objectstorage.StorableObject) {
	if otherReality, ok := other.(*Reality); !ok {
		panic("Update method expects a *TransferOutputBooking")
	} else {
		reality.parentRealityIdsMutex.Lock()
		reality.parentRealityIds = otherReality.parentRealityIds
		reality.parentRealityIdsMutex.Unlock()
	}
}

func (reality *Reality) MarshalBinary() ([]byte, error) {
	reality.parentRealityIdsMutex.RLock()

	parentRealityCount := len(reality.parentRealityIds)

	marshaledReality := make([]byte, 4+4+parentRealityCount*realityIdLength)

	binary.LittleEndian.PutUint32(marshaledReality, uint32(reality.GetTransferOutputCount()))

	binary.LittleEndian.PutUint32(marshaledReality[4:], uint32(parentRealityCount))
	i := 0
	for parentRealityId := range reality.parentRealityIds {
		copy(marshaledReality[4+4+i*realityIdLength:], parentRealityId[:])

		i++
	}

	reality.parentRealityIdsMutex.RUnlock()

	return marshaledReality, nil
}

func (reality *Reality) UnmarshalBinary(serializedObject []byte) error {
	if err := reality.id.UnmarshalBinary(reality.storageKey[:realityIdLength]); err != nil {
		return err
	}

	reality.parentRealityIds = NewRealityIdSet()

	reality.transferOutputCount = binary.LittleEndian.Uint32(serializedObject)

	parentRealityCount := int(binary.LittleEndian.Uint32(serializedObject[4:]))
	for i := 0; i < parentRealityCount; i++ {
		var restoredRealityId RealityId
		if err := restoredRealityId.UnmarshalBinary(serializedObject[4+4+i*realityIdLength:]); err != nil {
			return err
		}

		reality.parentRealityIds[restoredRealityId] = void
	}

	return nil
}
