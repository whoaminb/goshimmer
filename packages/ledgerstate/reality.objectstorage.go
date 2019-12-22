package ledgerstate

import (
	"encoding/binary"

	"github.com/iotaledger/goshimmer/packages/binary/types"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"

	"github.com/iotaledger/hive.go/bitmask"

	"github.com/iotaledger/hive.go/objectstorage"
)

func (mreality *Reality) GetStorageKey() []byte {
	return mreality.storageKey
}

func (mreality *Reality) Update(other objectstorage.StorableObject) {
	if otherReality, ok := other.(*Reality); !ok {
		panic("Update method expects a *TransferOutputBooking")
	} else {
		mreality.parentRealityIdsMutex.Lock()
		mreality.parentRealityIds = otherReality.parentRealityIds
		mreality.parentRealityIdsMutex.Unlock()
	}
}

func (mreality *Reality) MarshalBinary() ([]byte, error) {
	mreality.parentRealityIdsMutex.RLock()

	parentRealityCount := len(mreality.parentRealityIds)
	subRealityCount := len(mreality.subRealityIds)

	marshaledReality := make([]byte, 1+4+4+4+parentRealityCount*reality.IdLength+subRealityCount*reality.IdLength)

	offset := 0

	var flags bitmask.BitMask
	if mreality.IsPreferred() {
		flags = flags.SetFlag(0)
	}
	if mreality.IsLiked() {
		flags = flags.SetFlag(1)
	}
	marshaledReality[offset] = byte(flags)
	offset += 1

	binary.LittleEndian.PutUint32(marshaledReality[offset:], mreality.GetTransferOutputCount())
	offset += 4

	binary.LittleEndian.PutUint32(marshaledReality[offset:], uint32(parentRealityCount))
	offset += 4
	for parentRealityId := range mreality.parentRealityIds {
		copy(marshaledReality[offset:], parentRealityId[:])

		offset += reality.IdLength
	}

	binary.LittleEndian.PutUint32(marshaledReality[offset:], uint32(subRealityCount))
	offset += 4
	for subRealityId := range mreality.subRealityIds {
		copy(marshaledReality[offset:], subRealityId[:])

		offset += reality.IdLength
	}

	mreality.parentRealityIdsMutex.RUnlock()

	return marshaledReality, nil
}

func (mreality *Reality) UnmarshalBinary(serializedObject []byte) (err error) {
	if err = mreality.id.UnmarshalBinary(mreality.storageKey[:reality.IdLength]); err != nil {
		return
	}

	offset := 0

	mreality.unmarshalBinaryFlags(serializedObject, &offset)

	mreality.unmarshalBinaryTransferOutputCount(serializedObject, &offset)

	if err = mreality.unmarshalBinaryParentRealities(serializedObject, &offset); err != nil {
		return
	}

	if err = mreality.unmarshalBinarySubRealities(serializedObject, &offset); err != nil {
		return
	}

	return nil
}

func (mreality *Reality) unmarshalBinaryFlags(serializedObject []byte, offset *int) {
	var flags = bitmask.BitMask(serializedObject[*offset])

	if flags.HasFlag(0) {
		mreality.preferred = true
	}

	if flags.HasFlag(1) {
		mreality.liked = true
	}

	*offset += 1
}

func (mreality *Reality) unmarshalBinaryTransferOutputCount(serializedObject []byte, offset *int) {
	mreality.transferOutputCount = binary.LittleEndian.Uint32(serializedObject[*offset:])

	*offset = *offset + 4
}

func (mreality *Reality) unmarshalBinaryParentRealities(serializedObject []byte, offset *int) (err error) {
	mreality.parentRealityIds = reality.NewIdSet()

	parentRealityCount := int(binary.LittleEndian.Uint32(serializedObject[*offset:]))
	*offset += 4

	for i := 0; i < parentRealityCount; i++ {
		var restoredRealityId reality.Id
		if err = restoredRealityId.UnmarshalBinary(serializedObject[*offset:]); err != nil {
			return
		}
		*offset += reality.IdLength

		mreality.parentRealityIds[restoredRealityId] = types.Void
	}

	return
}

func (mreality *Reality) unmarshalBinarySubRealities(serializedObject []byte, offset *int) (err error) {
	mreality.subRealityIds = reality.NewIdSet()

	subRealityCount := int(binary.LittleEndian.Uint32(serializedObject[*offset:]))
	*offset += 4

	for i := 0; i < subRealityCount; i++ {
		var restoredRealityId reality.Id
		if err = restoredRealityId.UnmarshalBinary(serializedObject[*offset:]); err != nil {
			return
		}
		*offset += reality.IdLength

		mreality.subRealityIds[restoredRealityId] = types.Void
	}

	return
}
