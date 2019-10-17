package ram

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/interfaces"

	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
)

type TransferOutputStorage struct {
	id                     []byte
	unspentTransferOutputs map[hash.Reality]map[hash.Address]map[hash.Transfer]interfaces.TransferOutput
	spentTransferOutputs   map[hash.Reality]map[hash.Address]map[hash.Transfer]interfaces.TransferOutput
	mutex                  sync.RWMutex
}

func NewTransferOutputStorage(id []byte) interfaces.TransferOutputStorage {
	return &TransferOutputStorage{
		id:                     id,
		unspentTransferOutputs: make(map[hash.Reality]map[hash.Address]map[hash.Transfer]interfaces.TransferOutput),
		spentTransferOutputs:   make(map[hash.Reality]map[hash.Address]map[hash.Transfer]interfaces.TransferOutput),
	}
}

func (transferOutputStorage *TransferOutputStorage) StoreTransferOutput(transferOutput interfaces.TransferOutput) (err errors.IdentifiableError) {
	transferOutputStorage.mutex.Lock()

	var targetList map[hash.Reality]map[hash.Address]map[hash.Transfer]interfaces.TransferOutput
	if len(transferOutput.GetConsumers()) >= 1 {
		targetList = transferOutputStorage.spentTransferOutputs
	} else {
		targetList = transferOutputStorage.unspentTransferOutputs
	}

	reality, realityExists := targetList[transferOutput.GetRealityId()]
	if !realityExists {
		reality = make(map[hash.Address]map[hash.Transfer]interfaces.TransferOutput)

		targetList[transferOutput.GetRealityId()] = reality
	}

	address, addressExists := reality[transferOutput.GetAddressHash()]
	if !addressExists {
		address = make(map[hash.Transfer]interfaces.TransferOutput)

		reality[transferOutput.GetAddressHash()] = address
	}

	address[transferOutput.GetTransferHash()] = transferOutput

	transferOutputStorage.mutex.Unlock()

	return
}

func (transferOutputStorage *TransferOutputStorage) LoadTransferOutput(transferOutputReference interfaces.TransferOutputReference) (result interfaces.TransferOutput, err errors.IdentifiableError) {
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

func (transferOutputStorage *TransferOutputStorage) IterateRealities(realities map[hash.Reality]map[hash.Address]map[hash.Transfer]interfaces.TransferOutput, filter *interfaces.TransferOutputStorageFilters, callback func(transferOutput interfaces.TransferOutput)) {
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

func (transferOutputStorage *TransferOutputStorage) IterateAddresses(addresses map[hash.Address]map[hash.Transfer]interfaces.TransferOutput, filter *interfaces.TransferOutputStorageFilters, callback func(transferOutput interfaces.TransferOutput)) {
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

func (transferOutputStorage *TransferOutputStorage) IterateTransferOutputs(transferOutputs map[hash.Transfer]interfaces.TransferOutput, filter *interfaces.TransferOutputStorageFilters, callback func(transferOutput interfaces.TransferOutput)) {
	for _, transferOutput := range transferOutputs {
		callback(transferOutput)
	}
}

func (transferOutputStorage *TransferOutputStorage) ForEach(callback func(transferOutput interfaces.TransferOutput), filters ...interfaces.TransferOutputStorageFilter) {
	filter := interfaces.NewTransportOutputStorageFilter(filters...)

	if filter.FilterUnspent || !filter.FilterUnspent && !filter.FilterSpent {
		transferOutputStorage.IterateRealities(transferOutputStorage.unspentTransferOutputs, filter, callback)
	}
	if filter.FilterUnspent || !filter.FilterUnspent && !filter.FilterSpent {
		transferOutputStorage.IterateRealities(transferOutputStorage.spentTransferOutputs, filter, callback)
	}
}
