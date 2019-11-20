package ledgerstate

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/iotaledger/goshimmer/packages/errors"

	"github.com/iotaledger/goshimmer/packages/stringify"

	"github.com/iotaledger/hive.go/objectstorage"
)

type Reality struct {
	id              RealityId
	parentRealities map[RealityId]empty

	storageKey  []byte
	ledgerState *LedgerState

	parentRealitiesMutex sync.RWMutex
}

func newReality(id RealityId, parentRealities ...RealityId) *Reality {
	result := &Reality{
		id:              id,
		parentRealities: make(map[RealityId]empty),

		storageKey: make([]byte, len(id)),
	}
	copy(result.storageKey, id[:])

	for _, parentRealityId := range parentRealities {
		result.parentRealities[parentRealityId] = void
	}

	return result
}

func (reality *Reality) GetId() RealityId {
	return reality.id
}

func (reality *Reality) Elevate(oldParentRealityId RealityId, newParentRealityId RealityId) {
	reality.parentRealitiesMutex.Lock()

	fmt.Println(reality.id)

	if len(reality.parentRealities) > 1 {
		// aggregated reality
		fmt.Println("AGGREGATED REALITY")
		delete(reality.parentRealities, oldParentRealityId)
		reality.parentRealities[newParentRealityId] = void
	} else {
		delete(reality.parentRealities, oldParentRealityId)
		reality.parentRealities[newParentRealityId] = void
	}

	reality.parentRealitiesMutex.Unlock()
}

func (reality *Reality) DescendsFromReality(realityId RealityId) bool {
	if reality.id == realityId {
		return true
	} else {
		descendsFromReality := false

		for ancestorRealityId, ancestorReality := range reality.GetAncestorRealities() {
			if ancestorRealityId == realityId {
				descendsFromReality = true
			}

			ancestorReality.Release()
		}

		return descendsFromReality
	}
}

func (reality *Reality) GetParentRealities() map[RealityId]*objectstorage.CachedObject {
	parentRealities := make(map[RealityId]*objectstorage.CachedObject)

	reality.parentRealitiesMutex.RLock()

	for parentRealityId := range reality.parentRealities {
		loadedParentReality := reality.ledgerState.GetReality(parentRealityId)
		if !loadedParentReality.Exists() {
			reality.parentRealitiesMutex.RUnlock()

			panic("could not load parent reality with id \"" + string(parentRealityId[:]) + "\"")
		}

		parentRealities[loadedParentReality.Get().(*Reality).GetId()] = loadedParentReality
	}

	reality.parentRealitiesMutex.RUnlock()

	return parentRealities
}

func (reality *Reality) GetAncestorRealities() (result map[RealityId]*objectstorage.CachedObject) {
	result = make(map[RealityId]*objectstorage.CachedObject, 1)

	for parentRealityId, parentReality := range reality.GetParentRealities() {
		result[parentRealityId] = parentReality

		for ancestorId, ancestor := range parentReality.Get().(*Reality).GetAncestorRealities() {
			result[ancestorId] = ancestor
		}
	}

	return
}

func (reality *Reality) checkTransferBalances(inputs []*objectstorage.CachedObject, outputs map[AddressHash][]*ColoredBalance) error {
	totalColoredBalances := make(map[Color]uint64)

	for _, cachedInput := range inputs {
		if !cachedInput.Exists() {
			return errors.New("missing input in transfer")
		}

		transferOutput := cachedInput.Get().(*TransferOutput)
		if !reality.DescendsFromReality(transferOutput.GetRealityId()) {
			return errors.New("the referenced funds do not exist in this reality")
		}

		for _, balance := range transferOutput.GetBalances() {
			totalColoredBalances[balance.GetColor()] += balance.GetValue()
		}
	}

	for _, transferOutput := range outputs {
		for _, balance := range transferOutput {
			color := balance.GetColor()

			totalColoredBalances[color] -= balance.GetValue()

			if totalColoredBalances[color] == 0 {
				delete(totalColoredBalances, color)
			}
		}
	}

	// transfer is valid if sum of funds is 0
	if len(totalColoredBalances) != 0 {
		return errors.New("the sum of the balance changes is not 0")
	}

	return nil
}

func (reality *Reality) CreateReality(id RealityId) *objectstorage.CachedObject {
	newReality := newReality(id, reality.id)
	newReality.ledgerState = reality.ledgerState

	return reality.ledgerState.realities.Store(newReality)
}

func (reality *Reality) BookTransfer(transfer *Transfer) error {
	return reality.bookTransfer(transfer.GetHash(), reality.ledgerState.getTransferInputs(transfer), transfer.GetOutputs())
}

func (reality *Reality) elevateTransferOutput(transferOutputReference *TransferOutputReference, newRealityId RealityId) error {
	cachedTransferOutputToElevate := reality.ledgerState.GetTransferOutput(transferOutputReference)
	defer cachedTransferOutputToElevate.Release()

	if !cachedTransferOutputToElevate.Exists() {
		return errors.New("could not find TransferOutput to elevate")
	}

	transferOutputToElevate := cachedTransferOutputToElevate.Get().(*TransferOutput)
	if transferOutputToElevate.GetRealityId() == reality.id {
		if err := transferOutputToElevate.moveToReality(newRealityId); err != nil {
			return err
		}
		cachedTransferOutputToElevate.Store()

		for transferHash, addresses := range transferOutputToElevate.GetConsumers() {
			for _, addressHash := range addresses {
				if err := reality.elevateTransferOutput(NewTransferOutputReference(transferHash, addressHash), newRealityId); err != nil {
					return err
				}
			}
		}
	} else {
		reality.ledgerState.GetReality(transferOutputToElevate.GetRealityId()).Consume(func(nestedReality objectstorage.StorableObject) {
			nestedReality.(*Reality).Elevate(reality.id, newRealityId)
		})
	}

	return nil
}

// Creates a new reality for consumers that have previously been booked in this reality.
func (reality *Reality) elevateTransferOutputConsumersToOwnReality(consumers map[TransferHash][]AddressHash, conflictSet *objectstorage.CachedObject) {
	for transferHash, addressHashes := range consumers {
		var elevatedRealityId RealityId
		copy(elevatedRealityId[:], transferHash[:])
		reality.CreateReality(elevatedRealityId).Release()

		conflictSet.Get().(*ConflictSet).AddMember(elevatedRealityId)

		for _, addressHash := range addressHashes {
			reality.elevateTransferOutput(NewTransferOutputReference(transferHash, addressHash), elevatedRealityId)
		}
	}
}

func (reality *Reality) consumeInputs(inputs objectstorage.CachedObjects, transferHash TransferHash, outputs map[AddressHash][]*ColoredBalance) (conflictSets objectstorage.CachedObjects, err error) {
	conflictSets = make(objectstorage.CachedObjects, 0)

	for _, input := range inputs {
		consumedTransferOutput := input.Get().(*TransferOutput)

		inputConflicting, consumersToElevate, consumeErr := consumedTransferOutput.addConsumer(transferHash, outputs)
		if consumeErr != nil {
			err = consumeErr

			return
		}

		if inputConflicting {
			conflictSetId := NewConflictSetId(consumedTransferOutput.GetTransferHash(), consumedTransferOutput.GetAddressHash())

			var conflictSet *objectstorage.CachedObject
			if len(consumersToElevate) >= 1 {
				newConflictSet := newConflictSet(conflictSetId)
				newConflictSet.ledgerState = reality.ledgerState

				conflictSet = reality.ledgerState.conflictSets.Store(newConflictSet)

				reality.elevateTransferOutputConsumersToOwnReality(consumersToElevate, conflictSet)
			} else {
				conflictSet, err = reality.ledgerState.conflictSets.Load(conflictSetId[:])
				if err != nil {
					return
				}
				conflictSet.Get().(*ConflictSet).ledgerState = reality.ledgerState
			}

			conflictSets = append(conflictSets, conflictSet)
		}

		input.Store()
	}

	return
}

func (reality *Reality) bookTransfer(transferHash TransferHash, inputs objectstorage.CachedObjects, outputs map[AddressHash][]*ColoredBalance) error {
	if err := reality.checkTransferBalances(inputs, outputs); err != nil {
		return err
	}

	conflictSets, err := reality.consumeInputs(inputs, transferHash, outputs)
	if err != nil {
		return err
	}

	if len(conflictSets) >= 1 {
		var targetRealityId RealityId
		copy(targetRealityId[:], transferHash[:])

		for _, conflictSet := range conflictSets {
			conflictSet.Get().(*ConflictSet).AddMember(targetRealityId)
		}

		cachedTargetReality := reality.CreateReality(targetRealityId)
		cachedTargetReality.Get().(*Reality).bookTransferOutputs(transferHash, outputs)
		cachedTargetReality.Release()
	} else {
		reality.bookTransferOutputs(transferHash, outputs)
	}

	conflictSets.Release()
	inputs.Release()

	return nil
}

func (reality *Reality) bookTransferOutputs(transferHash TransferHash, transferOutputs map[AddressHash][]*ColoredBalance) {
	for addressHash, coloredBalances := range transferOutputs {
		createdTransferOutput := NewTransferOutput(reality.ledgerState, reality.id, transferHash, addressHash, coloredBalances...)
		createdBooking := newTransferOutputBooking(reality.id, addressHash, false, transferHash)

		reality.ledgerState.storeTransferOutput(createdTransferOutput).Release()
		reality.ledgerState.storeTransferOutputBooking(createdBooking).Release()
	}
}

func (reality *Reality) String() (result string) {
	reality.parentRealitiesMutex.RLock()

	parentRealities := make([]string, len(reality.parentRealities))

	i := 0
	for parentRealityId := range reality.parentRealities {
		parentRealities[i] = parentRealityId.String()

		i++
	}

	result = stringify.Struct("Reality",
		stringify.StructField("id", reality.id.String()),
		stringify.StructField("parentRealities", parentRealities),
	)

	reality.parentRealitiesMutex.RUnlock()

	return
}

// region support object storage ///////////////////////////////////////////////////////////////////////////////////////

func (reality *Reality) GetStorageKey() []byte {
	return reality.storageKey
}

func (reality *Reality) Update(other objectstorage.StorableObject) {
	reality.parentRealitiesMutex.Lock()

	if otherReality, ok := other.(*Reality); !ok {
		reality.parentRealitiesMutex.Unlock()

		panic("Update method expects a *TransferOutputBooking")
	} else {
		reality.parentRealities = otherReality.parentRealities
	}

	reality.parentRealitiesMutex.Unlock()
}

func (reality *Reality) MarshalBinary() ([]byte, error) {
	reality.parentRealitiesMutex.RLock()

	parentRealityCount := len(reality.parentRealities)

	marshaledReality := make([]byte, 4+parentRealityCount*realityIdLength)

	binary.LittleEndian.PutUint32(marshaledReality, uint32(parentRealityCount))
	i := 0
	for parentRealityId := range reality.parentRealities {
		copy(marshaledReality[4+i*realityIdLength:], parentRealityId[:])

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

	parentRealityCount := int(binary.LittleEndian.Uint32(serializedObject))
	for i := 0; i < parentRealityCount; i++ {
		var restoredRealityId RealityId
		if err := restoredRealityId.UnmarshalBinary(serializedObject[4+i*realityIdLength:]); err != nil {
			return err
		}

		reality.parentRealities[restoredRealityId] = void
	}

	return nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
