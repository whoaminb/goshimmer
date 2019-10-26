package ledgerstate

import (
	"encoding/binary"

	"github.com/iotaledger/goshimmer/packages/stringify"

	"github.com/iotaledger/goshimmer/packages/objectstorage"
)

type Reality struct {
	id              RealityId
	parentRealities []RealityId

	ledgerState *LedgerState
}

func newReality(id RealityId, parentRealities ...RealityId) *Reality {
	return &Reality{
		id:              id,
		parentRealities: parentRealities,
	}
}

func (reality *Reality) BookTransfer(transfer *Transfer) {
	// process outputs
	for addressHash, coloredBalances := range transfer.GetOutputs() {
		createdTransferOutput := NewTransferOutput(reality.ledgerState, reality.id, transfer.GetHash(), addressHash, coloredBalances...)
		reality.ledgerState.storeTransferOutput(createdTransferOutput).Release()
	}
}

func (reality *Reality) String() string {
	return stringify.Struct("Reality",
		stringify.StructField("id", reality.id.String()),
		stringify.StructField("parentRealities", reality.parentRealities),
	)
}

// region support object storage ///////////////////////////////////////////////////////////////////////////////////////

func (reality *Reality) GetId() []byte {
	return reality.id[:]
}

func (reality *Reality) Update(other objectstorage.StorableObject) {
	if otherReality, ok := other.(*Reality); !ok {
		panic("Update method expects a *TransferOutputBooking")
	} else {
		reality.parentRealities = otherReality.parentRealities
	}
}

func (reality *Reality) Marshal() ([]byte, error) {
	parentRealityCount := len(reality.parentRealities)

	marshaledReality := make([]byte, 4+parentRealityCount*realityIdLength)

	binary.LittleEndian.PutUint32(marshaledReality, uint32(parentRealityCount))
	for i := 0; i < parentRealityCount; i++ {
		copy(marshaledReality[4+i*realityIdLength:], reality.parentRealities[i][:])
	}

	return marshaledReality, nil
}

func (reality *Reality) Unmarshal(key []byte, serializedObject []byte) (objectstorage.StorableObject, error) {
	result := &Reality{}

	copy(result.id[:], key[:realityIdLength])

	parentRealityCount := int(binary.LittleEndian.Uint32(serializedObject))
	parentRealities := make([]RealityId, parentRealityCount)
	for i := 0; i < parentRealityCount; i++ {
		copy(parentRealities[i][:], serializedObject[4+i*realityIdLength:])
	}

	return result, nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
