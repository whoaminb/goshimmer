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
	subRealityCount := len(reality.subRealityIds)

	marshaledReality := make([]byte, 4+4+4+1+parentRealityCount*realityIdLength+subRealityCount*realityIdLength)

	offset := 0

	binary.LittleEndian.PutUint32(marshaledReality, uint32(reality.GetTransferOutputCount()))
	offset += 4

	binary.LittleEndian.PutUint32(marshaledReality[offset:], uint32(parentRealityCount))
	offset += 4
	for parentRealityId := range reality.parentRealityIds {
		copy(marshaledReality[offset:], parentRealityId[:])

		offset += realityIdLength
	}

	binary.LittleEndian.PutUint32(marshaledReality[offset:], uint32(subRealityCount))
	offset += 4
	for subRealityId := range reality.subRealityIds {
		copy(marshaledReality[offset:], subRealityId[:])

		offset += realityIdLength
	}

	if reality.liked {
		marshaledReality[offset] = 1
	} else {
		marshaledReality[offset] = 0
	}
	//offset += 1

	reality.parentRealityIdsMutex.RUnlock()

	return marshaledReality, nil
}

func (reality *Reality) UnmarshalBinary(serializedObject []byte) error {
	if err := reality.id.UnmarshalBinary(reality.storageKey[:realityIdLength]); err != nil {
		return err
	}

	reality.parentRealityIds = NewRealityIdSet()
	reality.subRealityIds = NewRealityIdSet()

	offset := 0

	reality.transferOutputCount = binary.LittleEndian.Uint32(serializedObject)
	offset += 4

	parentRealityCount := int(binary.LittleEndian.Uint32(serializedObject[offset:]))
	offset += 4

	for i := 0; i < parentRealityCount; i++ {
		var restoredRealityId RealityId
		if err := restoredRealityId.UnmarshalBinary(serializedObject[offset:]); err != nil {
			return err
		}
		offset += realityIdLength

		reality.parentRealityIds[restoredRealityId] = void
	}

	subRealityCount := int(binary.LittleEndian.Uint32(serializedObject[offset:]))
	offset += 4

	for i := 0; i < subRealityCount; i++ {
		var restoredRealityId RealityId
		if err := restoredRealityId.UnmarshalBinary(serializedObject[offset:]); err != nil {
			return err
		}
		offset += realityIdLength

		reality.subRealityIds[restoredRealityId] = void
	}

	reality.liked = serializedObject[offset] == 1
	//offset += 1

	return nil
}
