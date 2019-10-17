package ledgerstate

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/errors"
)

type TransferOutputStorageMemory struct {
	id                     []byte
	unspentTransferOutputs map[RealityId]map[AddressHash]map[TransferHash]*TransferOutput
	spentTransferOutputs   map[RealityId]map[AddressHash]map[TransferHash]*TransferOutput
	mutex                  sync.RWMutex
}

func newTransferOutputStorageMemory(id []byte) TransferOutputStorage {
	return &TransferOutputStorageMemory{
		id:                     id,
		unspentTransferOutputs: make(map[RealityId]map[AddressHash]map[TransferHash]*TransferOutput),
		spentTransferOutputs:   make(map[RealityId]map[AddressHash]map[TransferHash]*TransferOutput),
	}
}

func (transferOutputStorage *TransferOutputStorageMemory) StoreTransferOutput(transferOutput *TransferOutput) (err errors.IdentifiableError) {
	transferOutputStorage.mutex.Lock()

	var targetList map[RealityId]map[AddressHash]map[TransferHash]*TransferOutput
	if len(transferOutput.GetConsumers()) >= 1 {
		targetList = transferOutputStorage.spentTransferOutputs
	} else {
		targetList = transferOutputStorage.unspentTransferOutputs
	}

	reality, realityExists := targetList[transferOutput.GetRealityId()]
	if !realityExists {
		reality = make(map[AddressHash]map[TransferHash]*TransferOutput)

		targetList[transferOutput.GetRealityId()] = reality
	}

	address, addressExists := reality[transferOutput.GetAddressHash()]
	if !addressExists {
		address = make(map[TransferHash]*TransferOutput)

		reality[transferOutput.GetAddressHash()] = address
	}

	address[transferOutput.GetTransferHash()] = transferOutput

	transferOutputStorage.mutex.Unlock()

	return
}

func (transferOutputStorage *TransferOutputStorageMemory) LoadTransferOutput(transferOutputReference *TransferOutputReference) (result *TransferOutput, err errors.IdentifiableError) {
	transferOutputStorage.mutex.RLock()

	if reality, realityExists := transferOutputStorage.spentTransferOutputs[transferOutputReference.GetRealityId()]; realityExists {
		if address, addressExists := reality[transferOutputReference.GetAddressHash()]; addressExists {
			if transferOutput, transferOutputExists := address[transferOutputReference.GetTransferHash()]; transferOutputExists {
				result = transferOutput
			}
		}
	}

	if reality, realityExists := transferOutputStorage.unspentTransferOutputs[transferOutputReference.GetRealityId()]; realityExists {
		if address, addressExists := reality[transferOutputReference.GetAddressHash()]; addressExists {
			if transferOutput, transferOutputExists := address[transferOutputReference.GetTransferHash()]; transferOutputExists {
				result = transferOutput
			}
		}
	}

	transferOutputStorage.mutex.RUnlock()

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
	filter := newTransportOutputStorageFilters(filters...)

	if filter.FilterUnspent || !filter.FilterUnspent && !filter.FilterSpent {
		transferOutputStorage.IterateRealities(transferOutputStorage.unspentTransferOutputs, filter, callback)
	}
	if filter.FilterUnspent || !filter.FilterUnspent && !filter.FilterSpent {
		transferOutputStorage.IterateRealities(transferOutputStorage.spentTransferOutputs, filter, callback)
	}
}
