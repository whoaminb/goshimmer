package ledgerstate

import (
	"reflect"

	"github.com/iotaledger/goshimmer/packages/objectstorage"
)

type LedgerState struct {
	storageId              []byte
	transferOutputs        *objectstorage.ObjectStorage
	transferOutputBookings *objectstorage.ObjectStorage
	realities              *objectstorage.ObjectStorage
}

func NewLedgerState(storageId string) *LedgerState {
	return &LedgerState{
		storageId:              []byte(storageId),
		transferOutputs:        objectstorage.New(storageId+"TRANSFER_OUTPUTS", transferOutputFactory),
		transferOutputBookings: objectstorage.New(storageId+"TRANSFER_OUTPUT_BOOKING", transferOutputBookingFactory),
		realities:              objectstorage.New(storageId+"REALITIES", realityFactory),
	}
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
	ledgerState.realities.Store(newReality(id, MAIN_REALITY_ID))
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

					// TODO: FILTER UNSPENT + TRANSFER HASH
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
