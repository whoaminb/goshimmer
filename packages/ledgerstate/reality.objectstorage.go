package ledgerstate

import (
	"encoding/binary"

	"github.com/iotaledger/hive.go/bitmask"

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

	marshaledReality := make([]byte, 1+4+4+4+parentRealityCount*realityIdLength+subRealityCount*realityIdLength)

	offset := 0

	var flags bitmask.BitMask
	if reality.IsPreferred() {
		flags = flags.SetFlag(0)
	}
	if reality.IsLiked() {
		flags = flags.SetFlag(1)
	}
	marshaledReality[offset] = byte(flags)
	offset += 1

	binary.LittleEndian.PutUint32(marshaledReality[offset:], reality.GetTransferOutputCount())
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

	reality.parentRealityIdsMutex.RUnlock()

	return marshaledReality, nil
}

func (reality *Reality) UnmarshalBinary(serializedObject []byte) (err error) {
	if err = reality.id.UnmarshalBinary(reality.storageKey[:realityIdLength]); err != nil {
		return
	}

	offset := 0

	reality.unmarshalBinaryFlags(serializedObject, &offset)

	reality.unmarshalBinaryTransferOutputCount(serializedObject, &offset)

	if err = reality.unmarshalBinaryParentRealities(serializedObject, &offset); err != nil {
		return
	}

	if err = reality.unmarshalBinarySubRealities(serializedObject, &offset); err != nil {
		return
	}

	return nil
}

func (reality *Reality) unmarshalBinaryFlags(serializedObject []byte, offset *int) {
	var flags = bitmask.BitMask(serializedObject[*offset])

	if flags.HasFlag(0) {
		reality.preferred = true
	}

	if flags.HasFlag(1) {
		reality.liked = true
	}

	*offset += 1
}

func (reality *Reality) unmarshalBinaryTransferOutputCount(serializedObject []byte, offset *int) {
	reality.transferOutputCount = binary.LittleEndian.Uint32(serializedObject[*offset:])

	*offset = *offset + 4
}

func (reality *Reality) unmarshalBinaryParentRealities(serializedObject []byte, offset *int) (err error) {
	reality.parentRealityIds = NewRealityIdSet()

	parentRealityCount := int(binary.LittleEndian.Uint32(serializedObject[*offset:]))
	*offset += 4

	for i := 0; i < parentRealityCount; i++ {
		var restoredRealityId RealityId
		if err = restoredRealityId.UnmarshalBinary(serializedObject[*offset:]); err != nil {
			return
		}
		*offset += realityIdLength

		reality.parentRealityIds[restoredRealityId] = void
	}

	return
}

func (reality *Reality) unmarshalBinarySubRealities(serializedObject []byte, offset *int) (err error) {
	reality.subRealityIds = NewRealityIdSet()

	subRealityCount := int(binary.LittleEndian.Uint32(serializedObject[*offset:]))
	*offset += 4

	for i := 0; i < subRealityCount; i++ {
		var restoredRealityId RealityId
		if err = restoredRealityId.UnmarshalBinary(serializedObject[*offset:]); err != nil {
			return
		}
		*offset += realityIdLength

		reality.subRealityIds[restoredRealityId] = void
	}

	return
}
