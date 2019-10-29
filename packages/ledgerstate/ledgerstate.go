package ledgerstate

import (
	"reflect"
	"sort"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/objectstorage"
)

type LedgerState struct {
	storageId              []byte
	transferOutputs        *objectstorage.ObjectStorage
	transferOutputBookings *objectstorage.ObjectStorage
	realities              *objectstorage.ObjectStorage
}

func NewLedgerState(storageId string) *LedgerState {
	result := &LedgerState{
		storageId:              []byte(storageId),
		transferOutputs:        objectstorage.New(storageId+"TRANSFER_OUTPUTS", transferOutputFactory),
		transferOutputBookings: objectstorage.New(storageId+"TRANSFER_OUTPUT_BOOKING", transferOutputBookingFactory),
		realities:              objectstorage.New(storageId+"REALITIES", realityFactory),
	}

	mainReality := newReality(MAIN_REALITY_ID)
	mainReality.ledgerState = result
	result.realities.Store(mainReality).Release()

	return result
}

func (ledgerState *LedgerState) AddTransferOutput(transferHash TransferHash, addressHash AddressHash, balances ...*ColoredBalance) *LedgerState {
	ledgerState.storeTransferOutput(NewTransferOutput(ledgerState, MAIN_REALITY_ID, transferHash, addressHash, balances...)).Release()

	return ledgerState
}

func (ledgerState *LedgerState) storeTransferOutput(transferOutput *TransferOutput) *objectstorage.CachedObject {
	return ledgerState.transferOutputs.Store(transferOutput)
}

func (ledgerState *LedgerState) storeTransferOutputBooking(transferOutputBooking *TransferOutputBooking) *objectstorage.CachedObject {
	return ledgerState.transferOutputBookings.Store(transferOutputBooking)
}

func (ledgerState *LedgerState) GetTransferOutput(transferOutputReference *TransferOutputReference) *objectstorage.CachedObject {
	if cachedTransferOutput, err := ledgerState.transferOutputs.Load(transferOutputReference.GetStorageKey()); err != nil {
		panic(err)
	} else {
		if cachedTransferOutput.Exists() {
			if transferOutput := cachedTransferOutput.Get().(*TransferOutput); transferOutput != nil {
				transferOutput.ledgerState = ledgerState
			}
		}

		return cachedTransferOutput
	}
}

func (ledgerState *LedgerState) ForEachTransferOutput(callback func(object *objectstorage.CachedObject) bool, filters ...interface{}) {
	prefixes, searchBookings := ledgerState.generateFilterPrefixes(filters)
	if searchBookings {
		for _, prefix := range prefixes {
			if err := ledgerState.transferOutputBookings.ForEach(func(key []byte, cachedObject *objectstorage.CachedObject) bool {
				booking := cachedObject.Get().(*TransferOutputBooking)
				cachedObject.Release()

				return callback(ledgerState.GetTransferOutput(NewTransferOutputReference(booking.GetTransferHash(), booking.GetAddressHash())))
			}, prefix); err != nil {
				panic(err)
			}
		}
	} else {
		for _, prefix := range prefixes {
			if err := ledgerState.transferOutputs.ForEach(func(key []byte, cachedObject *objectstorage.CachedObject) bool {
				return callback(cachedObject)
			}, prefix); err != nil {
				panic(err)
			}
		}
	}
}

func (ledgerState *LedgerState) CreateReality(id RealityId) {
	newReality := newReality(id, MAIN_REALITY_ID)
	newReality.ledgerState = ledgerState

	ledgerState.realities.Store(newReality)
}

func (ledgerState *LedgerState) GetReality(id RealityId) *objectstorage.CachedObject {
	if cachedObject, err := ledgerState.realities.Load(id[:]); err != nil {
		panic(err)
	} else {
		if cachedObject.Exists() {
			if reality := cachedObject.Get().(*Reality); reality != nil {
				reality.ledgerState = ledgerState
			}
		}

		return cachedObject
	}
}

func (ledgerState *LedgerState) MergeRealities(realityIds ...RealityId) *objectstorage.CachedObject {
	switch len(realityIds) {
	case 0:
		if loadedReality, loadedRealityErr := ledgerState.realities.Load(MAIN_REALITY_ID[:]); loadedRealityErr != nil {
			panic(loadedRealityErr)
		} else {
			return loadedReality
		}
	case 1:
		if loadedReality, loadedRealityErr := ledgerState.realities.Load(realityIds[0][:]); loadedRealityErr != nil {
			panic(loadedRealityErr)
		} else {
			return loadedReality
		}
	default:
		aggregatedRealities := make(map[RealityId]*objectstorage.CachedObject)

	AGGREGATE_REALITIES:
		for _, realityId := range realityIds {
			// check if we have processed this reality already
			if _, exists := aggregatedRealities[realityId]; exists {
				continue
			}

			// load reality or abort if it fails
			cachedReality, loadingErr := ledgerState.realities.Load(realityId[:])
			if loadingErr != nil {
				panic(loadingErr)
			} else if !cachedReality.Exists() {
				panic(errors.New("referenced reality does not exist: " + realityId.String()))
			}

			// type cast the reality
			reality := cachedReality.Get().(*Reality)

			// check if the reality is already included in the aggregated realities
			for aggregatedRealityId, aggregatedReality := range aggregatedRealities {
				// if an already aggregated reality "descends" from the current reality, then we have found the more
				// "specialized" reality already and keep it
				if aggregatedReality.Get().(*Reality).DescendsFromReality(realityId) {
					continue AGGREGATE_REALITIES
				}

				// if the current reality
				if reality.DescendsFromReality(aggregatedRealityId) {
					delete(aggregatedRealities, aggregatedRealityId)
					aggregatedReality.Release()

					aggregatedRealities[reality.GetId()] = cachedReality

					continue AGGREGATE_REALITIES
				}
			}

			// store the reality as a new aggregated reality
			aggregatedRealities[realityId] = cachedReality
		}

		if len(aggregatedRealities) == 1 {
			for _, independentReality := range aggregatedRealities {
				return independentReality
			}
		}

		counter := 0
		aggregatedRealityIds := make([]RealityId, len(aggregatedRealities))
		for realityId, aggregatedReality := range aggregatedRealities {
			aggregatedRealityIds[counter] = realityId

			counter++

			aggregatedReality.Release()
		}

		sort.Slice(aggregatedRealityIds, func(i, j int) bool {
			for k := 0; k < len(aggregatedRealityIds[k]); k++ {
				if aggregatedRealityIds[i][k] < aggregatedRealityIds[j][k] {
					return true
				} else if aggregatedRealityIds[i][k] > aggregatedRealityIds[j][k] {
					return false
				}
			}

			return false
		})

		aggregatedRealityId := make([]byte, 0)
		for _, realityId := range aggregatedRealityIds {
			aggregatedRealityId = append(aggregatedRealityId, realityId[:]...)
		}

		aggregatedReality := newReality(blake2b.Sum256(aggregatedRealityId), aggregatedRealityIds...)
		aggregatedReality.ledgerState = ledgerState
		return ledgerState.realities.Prepare(aggregatedReality)
	}
}

func (ledgerState *LedgerState) Prune() *LedgerState {
	if err := ledgerState.transferOutputs.Prune(); err != nil {
		panic(err)
	}

	if err := ledgerState.transferOutputBookings.Prune(); err != nil {
		panic(err)
	}

	if err := ledgerState.realities.Prune(); err != nil {
		panic(err)
	}

	mainReality := newReality(MAIN_REALITY_ID)
	mainReality.ledgerState = ledgerState
	ledgerState.realities.Store(mainReality).Release()

	return ledgerState
}

func (ledgerState *LedgerState) generateFilterPrefixes(filters []interface{}) ([][]byte, bool) {
	filteredRealities := make([]RealityId, 0)
	filteredAddresses := make([]AddressHash, 0)
	filteredTransfers := make([]TransferHash, 0)
	filterSpent := false
	filterUnspent := false

	for _, filter := range filters {
		switch typeCastedValue := filter.(type) {
		case RealityId:
			filteredRealities = append(filteredRealities, typeCastedValue)
		case AddressHash:
			filteredAddresses = append(filteredAddresses, typeCastedValue)
		case TransferHash:
			filteredTransfers = append(filteredTransfers, typeCastedValue)
		case SpentIndicator:
			switch typeCastedValue {
			case SPENT:
				filterSpent = true
			case UNSPENT:
				filterUnspent = true
			default:
				panic("unknown SpentIndicator")
			}
		default:
			panic("unknown filter type: " + reflect.ValueOf(filter).Kind().String())
		}
	}

	prefixes := make([][]byte, 0)
	if len(filteredRealities) >= 1 {
		for _, realityId := range filteredRealities {
			realityPrefix := append([]byte{}, realityId[:]...)

			if len(filteredAddresses) >= 1 {
				for _, addressHash := range filteredAddresses {
					addressPrefix := append([]byte{}, realityPrefix...)
					addressPrefix = append(addressPrefix, addressHash[:]...)

					if filterSpent != filterUnspent {
						spentPrefix := append([]byte{}, addressPrefix...)
						if filterSpent {
							spentPrefix = append(spentPrefix, byte(SPENT))
						} else {
							spentPrefix = append(spentPrefix, byte(UNSPENT))
						}

						// TODO: FILTER TRANSFER HASH
						prefixes = append(prefixes, spentPrefix)
					} else {
						prefixes = append(prefixes, addressPrefix)
					}
				}
			} else {
				prefixes = append(prefixes, realityPrefix)
			}
		}

		return prefixes, true
	} else if len(filteredTransfers) >= 1 {
		for _, transferHash := range filteredTransfers {
			transferPrefix := append([]byte{}, transferHash[:]...)

			if len(filteredAddresses) >= 1 {
				for _, addressHash := range filteredAddresses {
					addressPrefix := append([]byte{}, transferPrefix...)
					addressPrefix = append(addressPrefix, addressHash[:]...)

					prefixes = append(prefixes, addressPrefix)
				}
			} else {
				prefixes = append(prefixes, transferPrefix)
			}
		}
	}

	return prefixes, false
}

func transferOutputFactory(key []byte) objectstorage.StorableObject {
	result := &TransferOutput{
		storageKey: make([]byte, len(key)),
	}
	copy(result.storageKey, key)

	return result
}

func transferOutputBookingFactory(key []byte) objectstorage.StorableObject {
	result := &TransferOutputBooking{
		storageKey: make([]byte, len(key)),
	}
	copy(result.storageKey, key)

	return result
}

func realityFactory(key []byte) objectstorage.StorableObject {
	result := &Reality{
		storageKey: make([]byte, len(key)),
	}
	copy(result.storageKey, key)

	return result
}
