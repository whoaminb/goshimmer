package ledgerstate

import (
	"sync"
	"sync/atomic"

	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/transfer"
	"github.com/iotaledger/goshimmer/packages/binary/types"
	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/conflict"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transferoutput"
	"github.com/iotaledger/goshimmer/packages/stringify"

	"github.com/iotaledger/hive.go/objectstorage"
)

type Reality struct {
	objectstorage.StorableObjectFlags

	id                    reality.Id
	parentRealityIds      reality.IdSet
	parentRealityIdsMutex sync.RWMutex
	subRealityIds         reality.IdSet
	subRealityIdsMutex    sync.RWMutex
	conflictIds           conflict.IdSet
	conflictIdsMutex      sync.RWMutex
	transferOutputCount   uint32
	preferred             bool
	preferredMutex        sync.RWMutex
	liked                 bool
	likedMutex            sync.RWMutex

	storageKey  []byte
	ledgerState *LedgerState
}

func (mreality *Reality) areParentsLiked() (parentsLiked bool) {
	parentsLiked = true
	for _, cachedParentReality := range mreality.GetParentRealities() {
		if parentsLiked {
			cachedParentReality.Consume(func(object objectstorage.StorableObject) {
				parentsLiked = parentsLiked && object.(*Reality).IsLiked()
			})
		} else {
			cachedParentReality.Release()
		}
	}

	return
}

func (mreality *Reality) propagateLiked() {
	mreality.likedMutex.Lock()
	mreality.liked = true
	mreality.likedMutex.Unlock()

	mreality.SetModified()

	for _, cachedSubReality := range mreality.GetSubRealities() {
		if !cachedSubReality.Exists() {
			cachedSubReality.Release()

			// TODO: SWITCH TO ERR INSTEAD OF PANIC
			panic("could not load sub reality")
		}

		cachedSubReality.Consume(func(object objectstorage.StorableObject) {
			subReality := object.(*Reality)

			subReality.parentRealityIdsMutex.RLock()
			if len(subReality.parentRealityIds) == 1 && subReality.parentRealityIds.Contains(mreality.id) {
				subReality.parentRealityIdsMutex.RUnlock()

				subReality.propagateLiked()
			} else {
				subReality.parentRealityIdsMutex.RUnlock()

				if subReality.areParentsLiked() {
					subReality.propagateLiked()
				}
			}
		})
	}
}

func (mreality *Reality) propagateDisliked() {
	mreality.likedMutex.Lock()
	mreality.liked = false
	mreality.likedMutex.Unlock()

	mreality.SetModified()

	for _, cachedSubReality := range mreality.GetSubRealities() {
		if !cachedSubReality.Exists() {
			cachedSubReality.Release()

			// TODO: SWITCH TO ERR INSTEAD OF PANIC
			panic("could not load sub reality")
		}

		cachedSubReality.Consume(func(object objectstorage.StorableObject) {
			subReality := object.(*Reality)

			if subReality.IsLiked() {
				subReality.propagateDisliked()
			}
		})
	}
}

func (mreality *Reality) GetSubRealities() (subRealities objectstorage.CachedObjects) {
	mreality.subRealityIdsMutex.RLock()
	subRealities = make(objectstorage.CachedObjects, len(mreality.subRealityIds))
	i := 0
	for subRealityId := range mreality.subRealityIds {
		subRealities[i] = mreality.ledgerState.GetReality(subRealityId)

		i++
	}
	mreality.subRealityIdsMutex.RUnlock()

	return
}

func (mreality *Reality) SetPreferred(preferred ...bool) (updated bool) {
	newPreferredValue := len(preferred) == 0 || preferred[0]

	mreality.preferredMutex.RLock()
	if mreality.preferred != newPreferredValue {
		mreality.preferredMutex.RUnlock()

		mreality.preferredMutex.Lock()
		if mreality.preferred != newPreferredValue {
			mreality.preferred = newPreferredValue

			if newPreferredValue {
				if mreality.areParentsLiked() {
					mreality.propagateLiked()
				}
			} else {
				if mreality.IsLiked() {
					mreality.propagateDisliked()
				}
			}

			updated = true

			mreality.SetModified()
		}
		mreality.preferredMutex.Unlock()
	} else {
		mreality.preferredMutex.RUnlock()
	}

	return
}

func (mreality *Reality) IsPreferred() (preferred bool) {
	mreality.preferredMutex.RLock()
	preferred = mreality.preferred
	mreality.preferredMutex.RUnlock()

	return
}

// region DONE REVIEWING ///////////////////////////////////////////////////////////////////////////////////////////////

// Creates a new Reality with the given id and parents. It is only used internally and therefore "private".
func newReality(id reality.Id, parentRealities ...reality.Id) *Reality {
	result := &Reality{
		id:               id,
		parentRealityIds: reality.NewIdSet(parentRealities...),
		subRealityIds:    reality.NewIdSet(),
		conflictIds:      conflict.NewIdSet(),

		storageKey: make([]byte, len(id)),
	}
	copy(result.storageKey, id[:])

	return result
}

func (mreality *Reality) IsLiked() (liked bool) {
	mreality.likedMutex.RLock()
	liked = mreality.liked
	mreality.likedMutex.RUnlock()

	return
}

func (mreality *Reality) SetLiked(liked ...bool) (likedStatusChanged bool) {
	newLikedStatus := len(liked) == 0 || liked[0]

	mreality.likedMutex.RLock()
	if mreality.liked != newLikedStatus {
		mreality.likedMutex.RUnlock()

		mreality.likedMutex.Lock()
		if mreality.liked != newLikedStatus {
			mreality.liked = newLikedStatus

			likedStatusChanged = true

			mreality.SetModified()
		}
		mreality.likedMutex.Unlock()
	} else {
		mreality.likedMutex.RUnlock()
	}

	return
}

// Returns the id of this Reality. Since the id never changes, we do not need a mutex to protect this property.
func (mreality *Reality) GetId() reality.Id {
	return mreality.id
}

// Returns the set of RealityIds that are the parents of this Reality (it creates a clone).
func (mreality *Reality) GetParentRealityIds() (realityIdSet reality.IdSet) {
	mreality.parentRealityIdsMutex.RLock()
	realityIdSet = mreality.parentRealityIds.Clone()
	mreality.parentRealityIdsMutex.RUnlock()

	return
}

// Adds a new parent Reality to this Reality (it is used for aggregating aggregated Realities).
func (mreality *Reality) AddParentReality(realityId reality.Id) (realityAdded bool) {
	mreality.parentRealityIdsMutex.RLock()
	if _, exists := mreality.parentRealityIds[realityId]; !exists {
		mreality.parentRealityIdsMutex.RUnlock()

		mreality.parentRealityIdsMutex.Lock()
		if _, exists := mreality.parentRealityIds[realityId]; !exists {
			mreality.parentRealityIds[realityId] = types.Void

			mreality.SetModified()

			realityAdded = true
		}
		mreality.parentRealityIdsMutex.Unlock()
	} else {
		mreality.parentRealityIdsMutex.RUnlock()
	}

	return
}

// Utility function that replaces the parent of a reality.
// Since IO is the most expensive part of the ledger state, we only update the parents and mark the reality as modified
// if either the oldRealityId exists or the newRealityId does not exist.
func (mreality *Reality) replaceParentReality(oldRealityId reality.Id, newRealityId reality.Id) {
	mreality.parentRealityIdsMutex.RLock()
	if _, oldRealityIdExist := mreality.parentRealityIds[oldRealityId]; oldRealityIdExist {
		mreality.parentRealityIdsMutex.RUnlock()

		mreality.parentRealityIdsMutex.Lock()
		if _, oldRealityIdExist := mreality.parentRealityIds[oldRealityId]; oldRealityIdExist {
			delete(mreality.parentRealityIds, oldRealityId)

			if _, newRealityIdExist := mreality.parentRealityIds[newRealityId]; !newRealityIdExist {
				mreality.parentRealityIds[newRealityId] = types.Void
			}

			mreality.SetModified()
		} else {
			if _, newRealityIdExist := mreality.parentRealityIds[newRealityId]; !newRealityIdExist {
				mreality.parentRealityIds[newRealityId] = types.Void

				mreality.SetModified()
			}
		}
		mreality.parentRealityIdsMutex.Unlock()
	} else {
		if _, newRealityIdExist := mreality.parentRealityIds[newRealityId]; !newRealityIdExist {
			mreality.parentRealityIdsMutex.RUnlock()

			mreality.parentRealityIdsMutex.Lock()
			if _, newRealityIdExist := mreality.parentRealityIds[newRealityId]; !newRealityIdExist {
				mreality.parentRealityIds[newRealityId] = types.Void

				mreality.SetModified()
			}
			mreality.parentRealityIdsMutex.Unlock()
		} else {
			mreality.parentRealityIdsMutex.RUnlock()
		}
	}
}

// Returns the amount of TransferOutputs in this Reality.
func (mreality *Reality) GetTransferOutputCount() uint32 {
	return atomic.LoadUint32(&(mreality.transferOutputCount))
}

// Increases (and returns) the amount of TransferOutputs in this Reality.
func (mreality *Reality) IncreaseTransferOutputCount() (transferOutputCount uint32) {
	transferOutputCount = atomic.AddUint32(&(mreality.transferOutputCount), 1)

	mreality.SetModified()

	return
}

// Decreases (and returns) the amount of TransferOutputs in this Reality.
func (mreality *Reality) DecreaseTransferOutputCount() (transferOutputCount uint32) {
	transferOutputCount = atomic.AddUint32(&(mreality.transferOutputCount), ^uint32(0))

	mreality.SetModified()

	return
}

// Returns true, if this reality is an "aggregated reality" that combines multiple other realities.
func (mreality *Reality) IsAggregated() (isAggregated bool) {
	mreality.parentRealityIdsMutex.RLock()
	isAggregated = len(mreality.parentRealityIds) > 1
	mreality.parentRealityIdsMutex.RUnlock()

	return
}

// Returns true if the given RealityId addresses the Reality itself or one of its ancestors.
func (mreality *Reality) DescendsFrom(realityId reality.Id) bool {
	if mreality.id == realityId {
		return true
	} else {
		descendsFromReality := false

		for ancestorRealityId, ancestorReality := range mreality.GetAncestorRealities() {
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
func (mreality *Reality) GetParentRealities() (parentRealities map[reality.Id]*objectstorage.CachedObject) {
	parentRealities = make(map[reality.Id]*objectstorage.CachedObject)

	mreality.parentRealityIdsMutex.RLock()
	for parentRealityId := range mreality.parentRealityIds {
		loadedParentReality := mreality.ledgerState.GetReality(parentRealityId)
		if !loadedParentReality.Exists() {
			mreality.parentRealityIdsMutex.RUnlock()

			panic("could not load parent reality with id \"" + string(parentRealityId[:]) + "\"")
		}

		parentRealities[loadedParentReality.Get().(*Reality).id] = loadedParentReality
	}
	mreality.parentRealityIdsMutex.RUnlock()

	return
}

// Returns a map of all parent realities that are not aggregated (aggregated realities are "transparent"). They have to
// be "released" manually when they are not needed anymore.
func (mreality *Reality) GetParentConflictRealities() map[reality.Id]*objectstorage.CachedObject {
	if !mreality.IsAggregated() {
		return mreality.GetParentRealities()
	} else {
		parentConflictRealities := make(map[reality.Id]*objectstorage.CachedObject)

		mreality.collectParentConflictRealities(parentConflictRealities)

		return parentConflictRealities
	}
}

// Returns a map of all ancestor realities (up till the MAIN_REALITY). They have to manually be "released" when they are
// not needed anymore.
func (mreality *Reality) GetAncestorRealities() (result map[reality.Id]*objectstorage.CachedObject) {
	result = make(map[reality.Id]*objectstorage.CachedObject, 1)

	for parentRealityId, parentReality := range mreality.GetParentRealities() {
		result[parentRealityId] = parentReality

		for ancestorId, ancestor := range parentReality.Get().(*Reality).GetAncestorRealities() {
			result[ancestorId] = ancestor
		}
	}

	return
}

// Registers the conflict set in the Reality.
func (mreality *Reality) AddConflict(conflictSetId conflict.Id) {
	mreality.conflictIdsMutex.RLock()
	if _, exists := mreality.conflictIds[conflictSetId]; !exists {
		mreality.conflictIdsMutex.RUnlock()

		mreality.conflictIdsMutex.Lock()
		if _, exists := mreality.conflictIds[conflictSetId]; !exists {
			mreality.conflictIds[conflictSetId] = types.Void

			mreality.SetModified()
		}
		mreality.conflictIdsMutex.Unlock()
	} else {
		mreality.conflictIdsMutex.RUnlock()
	}
}

// Creates a new sub Reality and "stores" it. It has to manually be "released" when it is not needed anymore.
func (mreality *Reality) CreateReality(id reality.Id) *objectstorage.CachedObject {
	newReality := newReality(id, mreality.id)
	newReality.ledgerState = mreality.ledgerState

	mreality.RegisterSubReality(id)

	return mreality.ledgerState.realities.Store(newReality)
}

// Books a transfer into this reality (wrapper for the private bookTransfer function).
func (mreality *Reality) BookTransfer(transfer *Transfer) (err error) {
	err = mreality.bookTransfer(transfer.GetHash(), mreality.ledgerState.getTransferInputs(transfer), transfer.GetOutputs())

	return
}

// Creates a string representation of this Reality.
func (mreality *Reality) String() (result string) {
	mreality.parentRealityIdsMutex.RLock()
	parentRealities := make([]string, len(mreality.parentRealityIds))
	i := 0
	for parentRealityId := range mreality.parentRealityIds {
		parentRealities[i] = parentRealityId.String()

		i++
	}
	mreality.parentRealityIdsMutex.RUnlock()

	result = stringify.Struct("Reality",
		stringify.StructField("id", mreality.GetId().String()),
		stringify.StructField("parentRealities", parentRealities),
	)

	return
}

// Books a transfer into this reality (contains the dispatcher for the actual tasks).
func (mreality *Reality) bookTransfer(transferHash transfer.Hash, inputs objectstorage.CachedObjects, outputs map[address.Address][]*ColoredBalance) (err error) {
	if err = mreality.verifyTransfer(inputs, outputs); err != nil {
		return
	}

	conflicts, err := mreality.consumeInputs(inputs, transferHash, outputs)
	if err != nil {
		return
	}

	if err = mreality.createTransferOutputs(transferHash, outputs, conflicts); err != nil {
		return
	}

	conflicts.Release()
	inputs.Release()

	return
}

// Internal utility function that verifies the transfer and checks if it is valid (inputs exist + the net balance is 0).
func (mreality *Reality) verifyTransfer(inputs []*objectstorage.CachedObject, outputs map[address.Address][]*ColoredBalance) error {
	totalColoredBalances := make(map[Color]uint64)

	for _, cachedInput := range inputs {
		if !cachedInput.Exists() {
			return errors.New("missing input in transfer")
		}

		input := cachedInput.Get().(*TransferOutput)
		if !mreality.DescendsFrom(input.GetRealityId()) {
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

// Internal utility function that marks the consumed inputs as spent and returns the corresponding conflicts if the
// inputs have been consumed before.
func (mreality *Reality) consumeInputs(inputs objectstorage.CachedObjects, transferHash transfer.Hash, outputs map[address.Address][]*ColoredBalance) (conflicts objectstorage.CachedObjects, err error) {
	conflicts = make(objectstorage.CachedObjects, 0)

	for _, input := range inputs {
		consumedInput := input.Get().(*TransferOutput)

		if consumersToElevate, consumeErr := consumedInput.addConsumer(transferHash, outputs); consumeErr != nil {
			err = consumeErr

			return
		} else if consumersToElevate != nil {
			if conflict, conflictErr := mreality.processConflictingInput(consumedInput, consumersToElevate); conflictErr != nil {
				err = conflictErr

				return
			} else {
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return
}

// Private utility function that creates the transfer outputs in the ledger.
//
// If the inputs have been used before and we consequently have a non-empty list of conflicts, we first create a new
// reality for the inputs and then book the transfer outputs into the correct reality.
func (mreality *Reality) createTransferOutputs(transferHash transfer.Hash, outputs map[address.Address][]*ColoredBalance, conflicts objectstorage.CachedObjects) (err error) {
	if len(conflicts) >= 1 {
		targetRealityId := transferHash.ToRealityId()

		mreality.CreateReality(targetRealityId).Consume(func(object objectstorage.StorableObject) {
			targetReality := object.(*Reality)

			for _, cachedConflictSet := range conflicts {
				conflictSet := cachedConflictSet.Get().(*conflict.Conflict)

				conflictSet.AddReality(targetRealityId)
				targetReality.AddConflict(conflictSet.GetId())
			}

			for addressHash, coloredBalances := range outputs {
				if err = targetReality.bookTransferOutput(NewTransferOutput(mreality.ledgerState, reality.EmptyId, transferHash, addressHash, coloredBalances...)); err != nil {
					return
				}
			}
		})
	} else {
		for addressHash, coloredBalances := range outputs {
			if err = mreality.bookTransferOutput(NewTransferOutput(mreality.ledgerState, reality.EmptyId, transferHash, addressHash, coloredBalances...)); err != nil {
				return
			}
		}
	}

	return
}

// Utility function that collects all non-aggregated parent realities. It is used by GetParentConflictRealities and
// prevents us from having to allocate multiple maps during recursion.
func (mreality *Reality) collectParentConflictRealities(parentConflictRealities map[reality.Id]*objectstorage.CachedObject) {
	for realityId, cachedParentReality := range mreality.GetParentRealities() {
		parentReality := cachedParentReality.Get().(*Reality)

		if !parentReality.IsAggregated() {
			parentConflictRealities[realityId] = cachedParentReality
		} else {
			parentReality.collectParentConflictRealities(parentConflictRealities)

			cachedParentReality.Release()
		}
	}
}

// Utility function that processes a conflicting input by retrieving the corresponding conflict.
// If there is a non-empty list of consumers to elevate, we elevate them.
func (mreality *Reality) processConflictingInput(input *TransferOutput, consumersToElevate map[transfer.Hash][]address.Address) (cachedConflict *objectstorage.CachedObject, err error) {
	conflictId := conflict.NewId(input.GetTransferHash(), input.GetAddressHash())

	if len(consumersToElevate) >= 1 {
		cachedConflict = mreality.ledgerState.conflictSets.Store(conflict.New(conflictId))

		err = mreality.createRealityForPreviouslyUnconflictingConsumers(consumersToElevate, cachedConflict.Get().(*conflict.Conflict))
	} else {
		if cachedConflict, err = mreality.ledgerState.conflictSets.Load(conflictId[:]); err != nil {
			return
		}
	}

	return
}

// Creates a Reality for the consumers of the conflicting inputs and registers it as part of the corresponding Conflict.
func (mreality *Reality) createRealityForPreviouslyUnconflictingConsumers(consumersOfConflictingInput map[transfer.Hash][]address.Address, conflict *conflict.Conflict) (err error) {
	for transferHash, addressHashes := range consumersOfConflictingInput {
		elevatedRealityId := transferHash.ToRealityId()

		// Retrieve the Reality for this Transfer or create one if no Reality exists, yet.
		var realityIsNew bool
		if cachedElevatedReality, realityErr := mreality.ledgerState.realities.ComputeIfAbsent(elevatedRealityId[:], func(key []byte) (object objectstorage.StorableObject, e error) {
			newReality := newReality(elevatedRealityId, mreality.id)
			newReality.ledgerState = mreality.ledgerState
			newReality.SetPreferred()

			mreality.RegisterSubReality(elevatedRealityId)

			newReality.Persist()
			newReality.SetModified()

			realityIsNew = true

			return newReality, nil
		}); realityErr != nil {
			err = realityErr
		} else {
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
						if err = mreality.elevateTransferOutput(transferoutput.NewTransferOutputReference(transferHash, addressHash), elevatedReality); err != nil {
							return
						}
					}
				}
			})
		}
	}

	return
}

// Private utility function that elevates a transfer output to the given reality.
func (mreality *Reality) elevateTransferOutput(transferOutputReference *transferoutput.Reference, newReality *Reality) (err error) {
	if cachedTransferOutputToElevate := mreality.ledgerState.GetTransferOutput(transferOutputReference); !cachedTransferOutputToElevate.Exists() {
		err = errors.New("could not find TransferOutput to elevate")
	} else {
		cachedTransferOutputToElevate.Consume(func(object objectstorage.StorableObject) {
			transferOutputToElevate := object.(*TransferOutput)

			if currentTransferOutputRealityId := transferOutputToElevate.GetRealityId(); currentTransferOutputRealityId == mreality.GetId() {
				err = mreality.elevateTransferOutputOfCurrentReality(transferOutputToElevate, newReality)
			} else if cachedNestedReality := mreality.ledgerState.GetReality(currentTransferOutputRealityId); !cachedNestedReality.Exists() {
				err = errors.New("could not find nested reality to elevate TransferOutput")
			} else {
				cachedNestedReality.Consume(func(nestedReality objectstorage.StorableObject) {
					err = nestedReality.(*Reality).elevateTransferOutputOfNestedReality(transferOutputToElevate, mreality.GetId(), newReality.GetId())
				})
			}
		})
	}

	return
}

// Private utility function that elevates the transfer output from the current reality to the new reality.
func (mreality *Reality) elevateTransferOutputOfCurrentReality(transferOutput *TransferOutput, newReality *Reality) (err error) {
	for transferHash, addresses := range transferOutput.GetConsumers() {
		for _, addressHash := range addresses {
			if elevateErr := mreality.elevateTransferOutput(transferoutput.NewTransferOutputReference(transferHash, addressHash), newReality); elevateErr != nil {
				err = elevateErr

				return
			}
		}
	}

	err = newReality.bookTransferOutput(transferOutput)

	return
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

func (mreality *Reality) GetSubRealityIdCount() (subRealityIdCount int) {
	mreality.subRealityIdsMutex.RLock()
	subRealityIdCount = len(mreality.subRealityIds)
	mreality.subRealityIdsMutex.RUnlock()

	return
}

func (mreality *Reality) UnregisterSubReality(realityId reality.Id) {
	mreality.subRealityIdsMutex.RLock()
	if _, subRealityIdExists := mreality.subRealityIds[realityId]; subRealityIdExists {
		mreality.subRealityIdsMutex.RUnlock()

		mreality.subRealityIdsMutex.Lock()
		if _, subRealityIdExists := mreality.subRealityIds[realityId]; subRealityIdExists {
			delete(mreality.subRealityIds, realityId)

			mreality.SetModified()
		}
		mreality.subRealityIdsMutex.Unlock()
	} else {
		mreality.subRealityIdsMutex.RUnlock()
	}
}

func (mreality *Reality) RegisterSubReality(realityId reality.Id) {
	mreality.subRealityIdsMutex.RLock()
	if _, subRealityIdExists := mreality.subRealityIds[realityId]; !subRealityIdExists {
		mreality.subRealityIdsMutex.RUnlock()

		mreality.subRealityIdsMutex.Lock()
		if _, subRealityIdExists := mreality.subRealityIds[realityId]; !subRealityIdExists {
			mreality.subRealityIds[realityId] = types.Void

			mreality.SetModified()
		}
		mreality.subRealityIdsMutex.Unlock()
	} else {
		mreality.subRealityIdsMutex.RUnlock()
	}
}

func (mreality *Reality) elevateTransferOutputOfNestedReality(transferOutput *TransferOutput, oldParentRealityId reality.Id, newParentRealityId reality.Id) (err error) {
	if !mreality.IsAggregated() {
		mreality.replaceParentReality(oldParentRealityId, newParentRealityId)
	} else {
		mreality.ledgerState.AggregateRealities(mreality.GetParentRealityIds().Remove(oldParentRealityId).Add(newParentRealityId).ToList()...).Consume(func(newAggregatedReality objectstorage.StorableObject) {
			newAggregatedReality.Persist()

			err = mreality.elevateTransferOutputOfCurrentReality(transferOutput, newAggregatedReality.(*Reality))
		})
	}

	return
}

func (mreality *Reality) bookTransferOutput(transferOutput *TransferOutput) (err error) {
	// retrieve required variables
	realityId := mreality.id
	transferOutputRealityId := transferOutput.GetRealityId()
	transferOutputAddressHash := transferOutput.GetAddressHash()
	transferOutputSpent := len(transferOutput.consumers) >= 1
	transferOutputTransferHash := transferOutput.GetTransferHash()

	// store the transferOutput if it is "new"
	if transferOutputRealityId == reality.EmptyId {
		transferOutput.SetRealityId(realityId)

		mreality.ledgerState.storeTransferOutput(transferOutput).Release()
	} else

	// remove old booking if the TransferOutput is currently booked in another reality
	if transferOutputRealityId != realityId {
		if oldTransferOutputBooking, err := mreality.ledgerState.transferOutputBookings.Load(generateTransferOutputBookingStorageKey(transferOutputRealityId, transferOutputAddressHash, len(transferOutput.consumers) >= 1, transferOutput.GetTransferHash())); err != nil {
			return err
		} else {
			transferOutput.SetRealityId(realityId)

			mreality.ledgerState.GetReality(transferOutputRealityId).Consume(func(object objectstorage.StorableObject) {
				transferOutputReality := object.(*Reality)

				// decrease transferOutputCount and remove reality if it is empty
				if transferOutputReality.DecreaseTransferOutputCount() == 0 && transferOutputReality.GetSubRealityIdCount() == 0 {
					for _, cachedParentReality := range transferOutputReality.GetParentRealities() {
						cachedParentReality.Consume(func(parentReality objectstorage.StorableObject) {
							parentReality.(*Reality).UnregisterSubReality(transferOutputRealityId)
						})
					}
					transferOutputReality.Delete()
				}
			})

			oldTransferOutputBooking.Consume(func(transferOutputBooking objectstorage.StorableObject) {
				transferOutputBooking.Delete()
			})
		}
	}

	// book the TransferOutput into the current Reality
	if transferOutputRealityId != realityId {
		mreality.ledgerState.storeTransferOutputBooking(newTransferOutputBooking(realityId, transferOutputAddressHash, transferOutputSpent, transferOutputTransferHash)).Release()

		mreality.IncreaseTransferOutputCount()
	}

	return
}
