package ledgerstate

import (
	"encoding/binary"

	"github.com/iotaledger/hive.go/objectstorage"
)

func (reality *Reality) GetStorageKey() []byte {
	return reality.storageKey
}

func (reality *Reality) Update(other objectstorage.StorableObject) {
	reality.bookingMutex.Lock()

	if otherReality, ok := other.(*Reality); !ok {
		reality.bookingMutex.Unlock()

		panic("Update method expects a *TransferOutputBooking")
	} else {
		reality.parentRealities = otherReality.parentRealities
	}

	reality.bookingMutex.Unlock()
}

func (reality *Reality) MarshalBinary() ([]byte, error) {
	reality.parentRealitiesMutex.RLock()

	parentRealityCount := len(reality.parentRealities)

	marshaledReality := make([]byte, 4+4+parentRealityCount*realityIdLength)

	binary.LittleEndian.PutUint32(marshaledReality, uint32(reality.transferOutputCount))

	binary.LittleEndian.PutUint32(marshaledReality[4:], uint32(parentRealityCount))
	i := 0
	for parentRealityId := range reality.parentRealities {
		copy(marshaledReality[4+4+i*realityIdLength:], parentRealityId[:])

		i++
	}

	reality.parentRealitiesMutex.RUnlock()

	return marshaledReality, nil
}

func (reality *Reality) UnmarshalBinary(serializedObject []byte) error {
	if err := reality.id.UnmarshalBinary(reality.storageKey[:realityIdLength]); err != nil {
		return err
	}

	reality.parentRealities = make(map[RealityId]empty)

	reality.transferOutputCount = int(binary.LittleEndian.Uint32(serializedObject))

	parentRealityCount := int(binary.LittleEndian.Uint32(serializedObject[4:]))
	for i := 0; i < parentRealityCount; i++ {
		var restoredRealityId RealityId
		if err := restoredRealityId.UnmarshalBinary(serializedObject[4+4+i*realityIdLength:]); err != nil {
			return err
		}

		reality.parentRealities[restoredRealityId] = void
	}

	return nil
}
