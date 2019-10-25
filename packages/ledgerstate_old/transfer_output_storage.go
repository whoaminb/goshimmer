package ledgerstate

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/objectstorage"

	"github.com/iotaledger/goshimmer/packages/errors"
)

// region TransferOutputStorage ////////////////////////////////////////////////////////////////////////////////////////

type TransferOutputStorage interface {
	LoadTransferOutput(transferOutputReference *TransferOutputReferenceOld) (result *TransferOutput, err errors.IdentifiableError)
	StoreTransferOutput(transferOutput *TransferOutput) (err errors.IdentifiableError)
	ForEach(callback func(transferOutput *TransferOutput), filters ...TransferOutputStorageFilter)
}

type TransferOutputStorageFactory func(id []byte) TransferOutputStorage

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region TransferOutputStorageFilters /////////////////////////////////////////////////////////////////////////////////

type TransferOutputStorageFilters struct {
	FilterUnspent bool
	FilterSpent   bool
	Realities     map[RealityId]bool
	Addresses     map[AddressHash]bool
}

func newTransportOutputStorageFilters(optionalFilters ...TransferOutputStorageFilter) *TransferOutputStorageFilters {
	result := &TransferOutputStorageFilters{
		FilterUnspent: false,
		FilterSpent:   false,
		Realities:     make(map[RealityId]bool),
		Addresses:     make(map[AddressHash]bool),
	}

	for _, optionalFilter := range optionalFilters {
		optionalFilter(result)
	}

	return result
}

type TransferOutputStorageFilter func(*TransferOutputStorageFilters)

func FilterSpent() TransferOutputStorageFilter {
	return func(args *TransferOutputStorageFilters) {
		args.FilterSpent = true
	}
}

func FilterUnspent() TransferOutputStorageFilter {
	return func(args *TransferOutputStorageFilters) {
		args.FilterUnspent = true
	}
}

func FilterRealities(realities ...RealityId) TransferOutputStorageFilter {
	return func(args *TransferOutputStorageFilters) {
		for _, reality := range realities {
			args.Realities[reality] = true
		}
	}
}

func FilterAddresses(addresses ...AddressHash) TransferOutputStorageFilter {
	return func(args *TransferOutputStorageFilters) {
		for _, reality := range addresses {
			args.Addresses[reality] = true
		}
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region TransportOutputStorageMemory /////////////////////////////////////////////////////////////////////////////////

type TransferOutputStorageMemory struct {
	id []byte
	// the actual transfer outputs (prefixed by reality, address, spent/unspent, transfer hash)
	transferOutputs        *objectstorage.ObjectStorage
	transferOutputBookings *objectstorage.ObjectStorage

	unspentTransferOutputsOld map[RealityId]map[AddressHash]map[TransferHash]bool
	spentTransferOutputsOld   map[RealityId]map[AddressHash]map[TransferHash]bool
	transferOutputsOld        map[AddressHash]map[TransferHash]*TransferOutput
	mutex                     sync.RWMutex
}

func newTransferOutputStorageMemory(id []byte) TransferOutputStorage {
	return &TransferOutputStorageMemory{
		id:                     id,
		transferOutputs:        objectstorage.New(string(id)+"TRANSFER_OUTPUTS", &TransferOutput{}),
		transferOutputBookings: objectstorage.New(string(id)+"TRANSFER_OUTPUT_REFERENCES", &TransferOutputBooking{}),

		unspentTransferOutputsOld: make(map[RealityId]map[AddressHash]map[TransferHash]bool),
		spentTransferOutputsOld:   make(map[RealityId]map[AddressHash]map[TransferHash]bool),
		transferOutputsOld:        make(map[AddressHash]map[TransferHash]*TransferOutput),
	}
}

func (transferOutputStorage *TransferOutputStorageMemory) StoreTransferOutput(transferOutput *TransferOutput) (err errors.IdentifiableError) {
	transferOutputStorage.transferOutputs.Store(transferOutput).Release()
	transferOutputStorage.mutex.Lock()

	var targetList map[RealityId]map[AddressHash]map[TransferHash]bool
	if len(transferOutput.GetConsumers()) >= 1 {
		targetList = transferOutputStorage.spentTransferOutputsOld
	} else {
		targetList = transferOutputStorage.unspentTransferOutputsOld
	}

	reality, realityExists := targetList[transferOutput.GetRealityId()]
	if !realityExists {
		reality = make(map[AddressHash]map[TransferHash]bool)

		targetList[transferOutput.GetRealityId()] = reality
	}

	address, addressExists := reality[transferOutput.GetAddressHash()]
	if !addressExists {
		address = make(map[TransferHash]bool)

		reality[transferOutput.GetAddressHash()] = address
	}

	address[transferOutput.GetTransferHash()] = true

	transferOutputStorage.mutex.Unlock()

	return
}

func (transferOutputStorage *TransferOutputStorageMemory) LoadTransferOutput(transferOutputReference *TransferOutputReferenceOld) (result *TransferOutput, err errors.IdentifiableError) {
	if cachedTransferOutput, loadErr := transferOutputStorage.transferOutputs.Load([]byte{1, 2}[:]); loadErr != nil {
		err = ErrInvalidTransfer.Derive(loadErr.Error())
	} else if cachedTransferOutput.Exists() {
		result = cachedTransferOutput.Get().(*TransferOutput)

		cachedTransferOutput.Release()
	}

	return
}

func (transferOutputStorage *TransferOutputStorageMemory) IterateRealities(realities map[RealityId]map[AddressHash]map[TransferHash]*TransferOutput, filter *TransferOutputStorageFilters, callback func(transferOutput *TransferOutput)) {
	if len(filter.Realities) >= 1 {
		for realityId := range filter.Realities {
			if reality, realityExists := realities[realityId]; realityExists {
				transferOutputStorage.IterateAddresses(reality, filter, callback)
			}
		}
	} else {
		for _, reality := range realities {
			transferOutputStorage.IterateAddresses(reality, filter, callback)
		}
	}
}

func (transferOutputStorage *TransferOutputStorageMemory) IterateAddresses(addresses map[AddressHash]map[TransferHash]*TransferOutput, filter *TransferOutputStorageFilters, callback func(transferOutput *TransferOutput)) {
	if len(filter.Addresses) >= 1 {
		for addressHash := range filter.Addresses {
			if address, addressExists := addresses[addressHash]; addressExists {
				transferOutputStorage.IterateTransferOutputs(address, filter, callback)
			}
		}
	} else {
		for _, address := range addresses {
			transferOutputStorage.IterateTransferOutputs(address, filter, callback)
		}
	}
}

func (transferOutputStorage *TransferOutputStorageMemory) IterateTransferOutputs(transferOutputs map[TransferHash]*TransferOutput, filter *TransferOutputStorageFilters, callback func(transferOutput *TransferOutput)) {
	for _, transferOutput := range transferOutputs {
		callback(transferOutput)
	}
}

func (transferOutputStorage *TransferOutputStorageMemory) ForEach(callback func(transferOutput *TransferOutput), filters ...TransferOutputStorageFilter) {
	//filter := newTransportOutputStorageFilters(filters...)

	/*
		if filter.FilterUnspent || !filter.FilterUnspent && !filter.FilterSpent {
			transferOutputStorage.IterateRealities(transferOutputStorage.unspentTransferOutputsOld, filter, callback)
		}
		if filter.FilterUnspent || !filter.FilterUnspent && !filter.FilterSpent {
			transferOutputStorage.IterateRealities(transferOutputStorage.spentTransferOutputsOld, filter, callback)
		}*/
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
