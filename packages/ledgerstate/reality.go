package ledgerstate

import (
	"fmt"
	"sync"

	"github.com/iotaledger/goshimmer/packages/errors"

	"github.com/iotaledger/goshimmer/packages/stringify"

	"github.com/iotaledger/hive.go/objectstorage"
)

type Reality struct {
	id                  RealityId
	transferOutputCount int
	parentRealities     map[RealityId]empty
	conflictSets        map[ConflictSetId]empty

	storageKey  []byte
	ledgerState *LedgerState

	bookingMutex             sync.RWMutex
	transferOutputCountMutex sync.RWMutex
	parentRealitiesMutex     sync.RWMutex
	conflictSetsMutex        sync.RWMutex
}

func newReality(id RealityId, parentRealities ...RealityId) *Reality {
	result := &Reality{
		id:              id,
		parentRealities: make(map[RealityId]empty),
		conflictSets:    make(map[ConflictSetId]empty),

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

func (reality *Reality) GetTransferOutputCount() (transferOutputCount int) {
	reality.transferOutputCountMutex.RLock()
	transferOutputCount = reality.transferOutputCount
	reality.transferOutputCountMutex.RUnlock()

	return
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

func (reality *Reality) AddConflictSet(conflictSetId ConflictSetId) {
	reality.conflictSetsMutex.Lock()
	reality.conflictSets[conflictSetId] = void
	reality.conflictSetsMutex.Unlock()
}

func (reality *Reality) CreateReality(id RealityId) *objectstorage.CachedObject {
	newReality := newReality(id, reality.id)
	newReality.ledgerState = reality.ledgerState

	return reality.ledgerState.realities.Store(newReality)
}

func (reality *Reality) BookTransfer(transfer *Transfer) (err error) {
	reality.bookingMutex.RLock()

	err = reality.bookTransfer(transfer.GetHash(), reality.ledgerState.getTransferInputs(transfer), transfer.GetOutputs())

	reality.bookingMutex.RUnlock()

	return
}

func (reality *Reality) bookTransfer(transferHash TransferHash, inputs objectstorage.CachedObjects, outputs map[AddressHash][]*ColoredBalance) error {
	if err := reality.verifyTransfer(inputs, outputs); err != nil {
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
			conflictSet.Get().(*ConflictSet).AddReality(targetRealityId)
		}

		reality.CreateReality(targetRealityId).Consume(func(object objectstorage.StorableObject) {
			object.(*Reality).persistTransfer(transferHash, outputs)
		})
	} else {
		reality.persistTransfer(transferHash, outputs)
	}

	conflictSets.Release()
	inputs.Release()

	return nil
}

// Verifies the transfer and checks if it is valid (spends existing funds + the net balance is 0).
func (reality *Reality) verifyTransfer(inputs []*objectstorage.CachedObject, outputs map[AddressHash][]*ColoredBalance) error {
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

// Marks the consumed inputs as spent and returns the corresponding ConflictSets if the inputs have been consumed before.
func (reality *Reality) consumeInputs(inputs objectstorage.CachedObjects, transferHash TransferHash, outputs map[AddressHash][]*ColoredBalance) (conflictSets objectstorage.CachedObjects, err error) {
	conflictSets = make(objectstorage.CachedObjects, 0)

	for _, input := range inputs {
		consumedTransferOutput := input.Get().(*TransferOutput)

		if consumersToElevate, consumeErr := consumedTransferOutput.addConsumer(transferHash, outputs); consumeErr != nil {
			err = consumeErr

			return
		} else if consumersToElevate != nil {
			if conflictSet, conflictErr := reality.retrieveConflictSetForConflictingInput(consumedTransferOutput, consumersToElevate); conflictErr != nil {
				err = conflictErr

				return
			} else {
				conflictSets = append(conflictSets, conflictSet)
			}
		}

		input.Store()
	}

	return
}

func (reality *Reality) retrieveConflictSetForConflictingInput(input *TransferOutput, consumersToElevate map[TransferHash][]AddressHash) (conflictSet *objectstorage.CachedObject, err error) {
	conflictSetId := NewConflictSetId(input.GetTransferHash(), input.GetAddressHash())

	if len(consumersToElevate) >= 1 {
		newConflictSet := newConflictSet(conflictSetId)
		newConflictSet.ledgerState = reality.ledgerState

		conflictSet = reality.ledgerState.conflictSets.Store(newConflictSet)

		err = reality.elevateTransferOutputs(consumersToElevate, conflictSet.Get().(*ConflictSet))
		if err != nil {
			return
		}
	} else {
		conflictSet, err = reality.ledgerState.conflictSets.Load(conflictSetId[:])
		if err != nil {
			return
		}
		conflictSet.Get().(*ConflictSet).ledgerState = reality.ledgerState
	}

	return
}

func (reality *Reality) elevateTransferOutputs(transferOutputs map[TransferHash][]AddressHash, conflictSet *ConflictSet) (err error) {
	for transferHash, addressHashes := range transferOutputs {
		// determine RealityId
		elevatedRealityId := transferHash.ToRealityId()

		// create new reality for every Transfer
		reality.CreateReality(elevatedRealityId).Consume(func(object objectstorage.StorableObject) {
			elevatedReality := object.(*Reality)

			// register Reality <-> ConflictSet
			conflictSet.AddReality(elevatedRealityId)
			elevatedReality.AddConflictSet(conflictSet.GetId())

			// elevate TransferOutputs
			for _, addressHash := range addressHashes {
				if err = reality.elevateTransferOutput(NewTransferOutputReference(transferHash, addressHash), elevatedReality); err != nil {
					return
				}
			}
		})
	}

	return
}

func (reality *Reality) elevateTransferOutput(transferOutputReference *TransferOutputReference, newReality *Reality) (err error) {
	if cachedTransferOutputToElevate := reality.ledgerState.GetTransferOutput(transferOutputReference); !cachedTransferOutputToElevate.Exists() {
		return errors.New("could not find TransferOutput to elevate")
	} else {
		cachedTransferOutputToElevate.Consume(func(object objectstorage.StorableObject) {
			transferOutputToElevate := object.(*TransferOutput)

			if transferOutputToElevate.GetRealityId() == reality.id {
				if moveErr := newReality.bookTransferOutput(transferOutputToElevate); moveErr != nil {
					err = moveErr

					return
				}

				for transferHash, addresses := range transferOutputToElevate.GetConsumers() {
					for _, addressHash := range addresses {
						if elevateErr := reality.elevateTransferOutput(NewTransferOutputReference(transferHash, addressHash), newReality); elevateErr != nil {
							err = elevateErr

							return
						}
					}
				}
			} else {
				reality.ledgerState.GetReality(transferOutputToElevate.GetRealityId()).Consume(func(nestedReality objectstorage.StorableObject) {
					nestedReality.(*Reality).elevateReality(reality.id, newReality.GetId())
				})
			}
		})
	}

	return
}

func (reality *Reality) elevateReality(oldParentRealityId RealityId, newParentRealityId RealityId) {
	reality.bookingMutex.Lock()
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
	reality.bookingMutex.Unlock()
}

func (reality *Reality) persistTransfer(transferHash TransferHash, transferOutputs map[AddressHash][]*ColoredBalance) {
	for addressHash, coloredBalances := range transferOutputs {
		reality.bookTransferOutput(NewTransferOutput(reality.ledgerState, emptyRealityId, transferHash, addressHash, coloredBalances...))
	}
}

func (reality *Reality) bookTransferOutput(transferOutput *TransferOutput) (err error) {
	// retrieve required variables
	realityId := reality.GetId()
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
				oldReality := object.(*Reality)

				oldReality.transferOutputCountMutex.Lock()
				oldReality.transferOutputCount--
				oldReality.transferOutputCountMutex.Unlock()
			})

			oldTransferOutputBooking.Delete().Release()
		}
	}

	// book the TransferOutput into the current Reality
	if transferOutputRealityId != realityId {
		reality.ledgerState.storeTransferOutputBooking(newTransferOutputBooking(realityId, transferOutputAddressHash, transferOutputSpent, transferOutputTransferHash)).Release()

		reality.transferOutputCountMutex.Lock()
		reality.transferOutputCount++
		reality.transferOutputCountMutex.Unlock()
	}

	return
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
