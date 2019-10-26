package ledgerstate

import (
	"encoding/binary"

	"github.com/iotaledger/goshimmer/packages/stringify"

	"github.com/iotaledger/goshimmer/packages/objectstorage"
)

type Reality struct {
	id              RealityId
	parentRealities []RealityId

	storageKey  []byte
	ledgerState *LedgerState
}

func newReality(id RealityId, parentRealities ...RealityId) *Reality {
	result := &Reality{
		id:              id,
		parentRealities: parentRealities,

		storageKey: make([]byte, len(id)),
	}
	copy(result.storageKey, id[:])

	return result
}

func (reality *Reality) BookTransfer(transfer *Transfer) {
	transferHash := transfer.GetHash()
	transferOutputs := transfer.GetOutputs()

	// process outputs
	reality.bookTransferOutputs(transferHash, transferOutputs)

}

func (reality *Reality) bookTransferOutputs(transferHash TransferHash, transferOutputs map[AddressHash][]*ColoredBalance) {
	for addressHash, coloredBalances := range transferOutputs {
		createdTransferOutput := NewTransferOutput(reality.ledgerState, reality.id, transferHash, addressHash, coloredBalances...)
		createdBooking := newTransferOutputBooking(reality.id, addressHash, false, transferHash)

		reality.ledgerState.storeTransferOutput(createdTransferOutput).Release()
		reality.ledgerState.storeTransferOutputBooking(createdBooking).Release()
	}
}

func (reality *Reality) String() string {
	return stringify.Struct("Reality",
		stringify.StructField("id", reality.id.String()),
		stringify.StructField("parentRealities", reality.parentRealities),
	)
}

// region support object storage ///////////////////////////////////////////////////////////////////////////////////////

func (reality *Reality) GetStorageKey() []byte {
	return reality.storageKey
}

func (reality *Reality) Update(other objectstorage.StorableObject) {
	if otherReality, ok := other.(*Reality); !ok {
		panic("Update method expects a *TransferOutputBooking")
	} else {
		reality.parentRealities = otherReality.parentRealities
	}
}

func (reality *Reality) MarshalBinary() ([]byte, error) {
	parentRealityCount := len(reality.parentRealities)

	marshaledReality := make([]byte, 4+parentRealityCount*realityIdLength)

	binary.LittleEndian.PutUint32(marshaledReality, uint32(parentRealityCount))
	for i := 0; i < parentRealityCount; i++ {
		copy(marshaledReality[4+i*realityIdLength:], reality.parentRealities[i][:])
	}

	return marshaledReality, nil
}

func (reality *Reality) UnmarshalBinary(serializedObject []byte) error {
	if err := reality.id.UnmarshalBinary(reality.storageKey[:realityIdLength]); err != nil {
		return err
	}

	parentRealityCount := int(binary.LittleEndian.Uint32(serializedObject))
	parentRealities := make([]RealityId, parentRealityCount)
	for i := 0; i < parentRealityCount; i++ {
		if err := parentRealities[i].UnmarshalBinary(serializedObject[4+i*realityIdLength:]); err != nil {
			return err
		}
	}

	return nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
