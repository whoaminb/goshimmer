package ledgerstate

import (
	"sync"
	"sync/atomic"

	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/stringify"

	"github.com/iotaledger/hive.go/objectstorage"
)

type Reality struct {
	objectstorage.StorableObjectFlags

	id                    RealityId
	parentRealityIds      RealityIdSet
	parentRealityIdsMutex sync.RWMutex
	conflictIds           ConflictIdSet
	conflictIdsMutex      sync.RWMutex
	transferOutputCount   uint32

	storageKey  []byte
	ledgerState *LedgerState
}

// region DONE REVIEWING ///////////////////////////////////////////////////////////////////////////////////////////////

// Creates a new Reality with the given id and parents. It is only used internally and therefore "private".
func newReality(id RealityId, parentRealities ...RealityId) *Reality {
	result := &Reality{
		id:               id,
		parentRealityIds: NewRealityIdSet(parentRealities...),
		conflictIds:      NewConflictIdSet(),

		storageKey: make([]byte, len(id)),
	}
	copy(result.storageKey, id[:])

	return result
}

// Returns the id of this Reality. Since the id never changes, we do not need a mutex to protect this property.
func (reality *Reality) GetId() RealityId {
	return reality.id
}

// Returns the set of RealityIds that are the parents of this Reality (it creates a clone).
func (reality *Reality) GetParentRealityIds() (realityIdSet RealityIdSet) {
	reality.parentRealityIdsMutex.RLock()
	realityIdSet = reality.parentRealityIds.Clone()
	reality.parentRealityIdsMutex.RUnlock()

	return
}

// Adds a new parent Reality to this Reality (it is used for aggregating aggregated Realities).
func (reality *Reality) AddParentReality(realityId RealityId) {
	reality.parentRealityIdsMutex.RLock()
	if _, exists := reality.parentRealityIds[realityId]; !exists {
		reality.parentRealityIdsMutex.RUnlock()

		reality.parentRealityIdsMutex.Lock()
		if _, exists := reality.parentRealityIds[realityId]; !exists {
			reality.parentRealityIds[realityId] = void

			reality.SetModified()
		}
		reality.parentRealityIdsMutex.Unlock()
	} else {
		reality.parentRealityIdsMutex.RUnlock()
	}
}

// Returns the amount of TransferOutputs in this Reality.
func (reality *Reality) GetTransferOutputCount() uint32 {
	return atomic.LoadUint32(&(reality.transferOutputCount))
}

// Increases (and returns) the amount of TransferOutputs in this Reality.
func (reality *Reality) IncreaseTransferOutputCount() uint32 {
	return atomic.AddUint32(&(reality.transferOutputCount), 1)
}

// Decreases (and returns) the amount of TransferOutputs in this Reality.
func (reality *Reality) DecreaseTransferOutputCount() uint32 {
	return atomic.AddUint32(&(reality.transferOutputCount), ^uint32(0))
}

// Returns true, if this reality is an "aggregated reality" that combines multiple other realities.
func (reality *Reality) IsAggregated() (isAggregated bool) {
	reality.parentRealityIdsMutex.RLock()
	isAggregated = len(reality.parentRealityIds) > 1
	reality.parentRealityIdsMutex.RUnlock()

	return
}

// Returns true if the given RealityId addresses the Reality itself or one of its ancestors.
func (reality *Reality) DescendsFrom(realityId RealityId) bool {
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

// Returns a map of all parent realities (one level). They have to be "released" manually when they are not needed
// anymore.
func (reality *Reality) GetParentRealities() (parentRealities map[RealityId]*objectstorage.CachedObject) {
	parentRealities = make(map[RealityId]*objectstorage.CachedObject)

	reality.parentRealityIdsMutex.RLock()
	for parentRealityId := range reality.parentRealityIds {
		loadedParentReality := reality.ledgerState.GetReality(parentRealityId)
		if !loadedParentReality.Exists() {
			reality.parentRealityIdsMutex.RUnlock()

			panic("could not load parent reality with id \"" + string(parentRealityId[:]) + "\"")
		}

		parentRealities[loadedParentReality.Get().(*Reality).id] = loadedParentReality
	}
	reality.parentRealityIdsMutex.RUnlock()

	return
}

// Returns a map of all parent realities that are conflicting. Aggregated realities are "transparent". They have to be
// "released" manually when they are not needed anymore.
func (reality *Reality) GetParentConflictRealities() map[RealityId]*objectstorage.CachedObject {
	if !reality.IsAggregated() {
		return reality.GetParentRealities()
	} else {
		parentConflictRealities := make(map[RealityId]*objectstorage.CachedObject)

		reality.collectParentConflictRealities(parentConflictRealities)

		return parentConflictRealities
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

func (reality *Reality) collectParentConflictRealities(parentConflictRealities map[RealityId]*objectstorage.CachedObject) {
	for realityId, cachedParentReality := range reality.GetParentRealities() {
		parentReality := cachedParentReality.Get().(*Reality)

		if !parentReality.IsAggregated() {
			parentConflictRealities[realityId] = cachedParentReality
		} else {
			parentReality.collectParentConflictRealities(parentConflictRealities)

			cachedParentReality.Release()
		}
	}
}

// [DONE] Returns a map of all ancestor realities (up till the MAIN_REALITY). They have to manually be "released" when
// they are not needed anymore.
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

// [DONE] Registers the conflict set in the Reality.
func (reality *Reality) AddConflict(conflictSetId ConflictId) {
	reality.conflictIdsMutex.Lock()
	if _, exists := reality.conflictIds[conflictSetId]; !exists {
		reality.conflictIds[conflictSetId] = void

		reality.SetModified()
	}
	reality.conflictIdsMutex.Unlock()
}

// [DONE] Creates a new sub Reality and "stores" it. It has to manually be "released" when it is not needed anymore.
func (reality *Reality) CreateReality(id RealityId) *objectstorage.CachedObject {
	newReality := newReality(id, reality.id)
	newReality.ledgerState = reality.ledgerState

	return reality.ledgerState.realities.Store(newReality)
}

func (reality *Reality) BookTransfer(transfer *Transfer) (err error) {
	err = reality.bookTransfer(transfer.GetHash(), reality.ledgerState.getTransferInputs(transfer), transfer.GetOutputs())

	return
}

func (reality *Reality) bookTransfer(transferHash TransferHash, inputs objectstorage.CachedObjects, outputs map[AddressHash][]*ColoredBalance) (err error) {
	if err = reality.verifyTransfer(inputs, outputs); err != nil {
		return
	}

	conflicts, err := reality.consumeInputs(inputs, transferHash, outputs)
	if err != nil {
		return
	}

	if len(conflicts) >= 1 {
		targetRealityId := transferHash.ToRealityId()

		reality.CreateReality(targetRealityId).Consume(func(object objectstorage.StorableObject) {
			targetReality := object.(*Reality)

			for _, cachedConflictSet := range conflicts {
				conflictSet := cachedConflictSet.Get().(*Conflict)

				conflictSet.AddReality(targetRealityId)
				targetReality.AddConflict(conflictSet.GetId())
			}

			for addressHash, coloredBalances := range outputs {
				if err = targetReality.bookTransferOutput(NewTransferOutput(reality.ledgerState, emptyRealityId, transferHash, addressHash, coloredBalances...)); err != nil {
					return
				}
			}
		})
	} else {
		for addressHash, coloredBalances := range outputs {
			if err = reality.bookTransferOutput(NewTransferOutput(reality.ledgerState, emptyRealityId, transferHash, addressHash, coloredBalances...)); err != nil {
				return
			}
		}
	}

	conflicts.Release()
	inputs.Release()

	return
}

// Verifies the transfer and checks if it is valid (spends existing funds + the net balance is 0).
func (reality *Reality) verifyTransfer(inputs []*objectstorage.CachedObject, outputs map[AddressHash][]*ColoredBalance) error {
	totalColoredBalances := make(map[Color]uint64)

	for _, cachedInput := range inputs {
		if !cachedInput.Exists() {
			return errors.New("missing input in transfer")
		}

		input := cachedInput.Get().(*TransferOutput)
		if !reality.DescendsFrom(input.GetRealityId()) {
			return errors.New("the referenced funds do not exist in this reality")
		}

		for _, balance := range input.GetBalances() {
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

// Marks the consumed inputs as spent and returns the corresponding Conflict if the inputs have been consumed before.
func (reality *Reality) consumeInputs(inputs objectstorage.CachedObjects, transferHash TransferHash, outputs map[AddressHash][]*ColoredBalance) (conflicts objectstorage.CachedObjects, err error) {
	conflicts = make(objectstorage.CachedObjects, 0)

	for _, input := range inputs {
		consumedInput := input.Get().(*TransferOutput)

		if consumersToElevate, consumeErr := consumedInput.addConsumer(transferHash, outputs); consumeErr != nil {
			err = consumeErr

			return
		} else if consumersToElevate != nil {
			if conflict, conflictErr := reality.retrieveConflictForConflictingInput(consumedInput, consumersToElevate); conflictErr != nil {
				err = conflictErr

				return
			} else {
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return
}

func (reality *Reality) retrieveConflictForConflictingInput(input *TransferOutput, consumersToElevate map[TransferHash][]AddressHash) (conflict *objectstorage.CachedObject, err error) {
	conflictSetId := NewConflictSetId(input.GetTransferHash(), input.GetAddressHash())

	if len(consumersToElevate) >= 1 {
		newConflict := newConflictSet(conflictSetId)
		newConflict.ledgerState = reality.ledgerState

		conflict = reality.ledgerState.conflictSets.Store(newConflict)

		err = reality.createRealityForConsumerOfConflictingInput(consumersToElevate, conflict.Get().(*Conflict))
	} else {
		if conflict, err = reality.ledgerState.conflictSets.Load(conflictSetId[:]); err == nil {
			conflict.Get().(*Conflict).ledgerState = reality.ledgerState
		}
	}

	return
}

// Creates a Reality for the consumers of the conflicting inputs and registers it as part of the corresponding Conflict.
func (reality *Reality) createRealityForConsumerOfConflictingInput(consumersOfConflictingInput map[TransferHash][]AddressHash, conflict *Conflict) (err error) {
	for transferHash, addressHashes := range consumersOfConflictingInput {
		var elevatedRealityId = transferHash.ToRealityId()
		var realityIsNew bool
		var cachedElevatedReality *objectstorage.CachedObject

		// Retrieve the Reality for this Transfer or create one if no Reality exists, yet.
		if cachedElevatedReality, err = reality.ledgerState.realities.ComputeIfAbsent(elevatedRealityId[:], func(key []byte) (object objectstorage.StorableObject, e error) {
			newReality := newReality(elevatedRealityId, reality.id)
			newReality.ledgerState = reality.ledgerState

			newReality.Persist()
			newReality.SetModified()

			realityIsNew = true

			return newReality, nil
		}); err == nil {
			cachedElevatedReality.Consume(func(object objectstorage.StorableObject) {
				elevatedReality := object.(*Reality)

				// We register every Conflict with the Reality (independent if it is "new" or not), to reflect its
				// association to all corresponding Conflicts. (Note: A Reality can be part of multiple Conflicts if the
				// Transfer that is associated to this Reality consumes multiple inputs.
				conflict.AddReality(elevatedRealityId)
				elevatedReality.AddConflict(conflict.GetId())

				// A transaction can consume multiple inputs. We only elevate the consumers of a Reality once (when the
				// Reality is created the first time).
				if realityIsNew {
					for _, addressHash := range addressHashes {
						if err = reality.elevateTransferOutput(NewTransferOutputReference(transferHash, addressHash), elevatedReality); err != nil {
							return
						}
					}
				}
			})
		}
	}

	return
}

func (reality *Reality) elevateTransferOutput(transferOutputReference *TransferOutputReference, newReality *Reality) (err error) {
	cachedTransferOutputToElevate := reality.ledgerState.GetTransferOutput(transferOutputReference)
	if !cachedTransferOutputToElevate.Exists() {
		return errors.New("could not find TransferOutput to elevate")
	}

	cachedTransferOutputToElevate.Consume(func(object objectstorage.StorableObject) {
		transferOutputToElevate := object.(*TransferOutput)

		transferOutputToElevate.SetModified()

		if transferOutputToElevate.GetRealityId() == reality.id {
			err = reality.elevateTransferOutputOfCurrentReality(transferOutputToElevate, newReality)
		} else {
			reality.ledgerState.GetReality(transferOutputToElevate.GetRealityId()).Consume(func(nestedReality objectstorage.StorableObject) {
				err = nestedReality.(*Reality).elevateTransferOutputOfNestedReality(transferOutputToElevate, reality.id, newReality.id)
			})
		}
	})

	return
}

func (reality *Reality) elevateTransferOutputOfCurrentReality(transferOutput *TransferOutput, newReality *Reality) (err error) {
	if err = newReality.bookTransferOutput(transferOutput); err != nil {
		return
	}

	for transferHash, addresses := range transferOutput.GetConsumers() {
		for _, addressHash := range addresses {
			if elevateErr := reality.elevateTransferOutput(NewTransferOutputReference(transferHash, addressHash), newReality); elevateErr != nil {
				err = elevateErr

				return
			}
		}
	}

	return
}

func (reality *Reality) elevateTransferOutputOfNestedReality(transferOutput *TransferOutput, oldParentRealityId RealityId, newParentRealityId RealityId) (err error) {
	if !reality.IsAggregated() {
		reality.parentRealityIdsMutex.Lock()
		reality.parentRealityIds.Remove(oldParentRealityId).Add(newParentRealityId)
		reality.parentRealityIdsMutex.Unlock()

		return
	}

	newParentRealities := reality.GetParentRealityIds().Remove(oldParentRealityId).Add(newParentRealityId).ToList()

	reality.ledgerState.AggregateRealities(newParentRealities...).Consume(func(object objectstorage.StorableObject) {
		object.Persist()
		object.SetModified()

		err = reality.elevateTransferOutputOfCurrentReality(transferOutput, object.(*Reality))
	})

	return
}

func (reality *Reality) bookTransferOutput(transferOutput *TransferOutput) (err error) {
	// retrieve required variables
	realityId := reality.id
	transferOutputRealityId := transferOutput.GetRealityId()
	transferOutputAddressHash := transferOutput.GetAddressHash()
	transferOutputSpent := len(transferOutput.consumers) >= 1
	transferOutputTransferHash := transferOutput.GetTransferHash()

	// store the transferOutput if it is "new"
	if transferOutputRealityId == emptyRealityId {
		transferOutput.SetRealityId(realityId)

		reality.ledgerState.storeTransferOutput(transferOutput).Release()
	} else

	// remove old booking if the TransferOutput is currently booked in another reality
	if transferOutputRealityId != realityId {
		if oldTransferOutputBooking, err := reality.ledgerState.transferOutputBookings.Load(generateTransferOutputBookingStorageKey(transferOutputRealityId, transferOutputAddressHash, len(transferOutput.consumers) >= 1, transferOutput.GetTransferHash())); err != nil {
			return err
		} else {
			transferOutput.SetRealityId(realityId)

			reality.ledgerState.GetReality(transferOutputRealityId).Consume(func(object objectstorage.StorableObject) {
				// decrease transferOutputCount and remove reality if it is empty
				if object.(*Reality).DecreaseTransferOutputCount() == 0 {
					//reality.ledgerState.realities.Delete(transferOutputRealityId[:])
				}
			})

			oldTransferOutputBooking.Consume(func(transferOutputBooking objectstorage.StorableObject) {
				transferOutputBooking.Delete()
			})
		}
	}

	// book the TransferOutput into the current Reality
	if transferOutputRealityId != realityId {
		reality.ledgerState.storeTransferOutputBooking(newTransferOutputBooking(realityId, transferOutputAddressHash, transferOutputSpent, transferOutputTransferHash)).Release()

		reality.IncreaseTransferOutputCount()
	}

	return
}

func (reality *Reality) String() (result string) {
	reality.parentRealityIdsMutex.RLock()

	parentRealities := make([]string, len(reality.parentRealityIds))
	i := 0
	for parentRealityId := range reality.parentRealityIds {
		parentRealities[i] = parentRealityId.String()

		i++
	}

	result = stringify.Struct("Reality",
		stringify.StructField("id", reality.id.String()),
		stringify.StructField("parentRealities", parentRealities),
	)

	reality.parentRealityIdsMutex.RUnlock()

	return
}
