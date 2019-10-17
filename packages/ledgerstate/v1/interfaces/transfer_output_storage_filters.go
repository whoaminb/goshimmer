package interfaces

import "github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"

func NewTransportOutputStorageFilter(optionalFilters ...TransferOutputStorageFilter) *TransferOutputStorageFilters {
	result := &TransferOutputStorageFilters{
		FilterUnspent: false,
		FilterSpent:   false,
		Realities:     make(map[hash.Reality]bool),
		Addresses:     make(map[hash.Address]bool),
	}

	for _, optionalFilter := range optionalFilters {
		optionalFilter(result)
	}

	return result
}

type TransferOutputStorageFilter func(*TransferOutputStorageFilters)

type TransferOutputStorageFilters struct {
	FilterUnspent bool
	FilterSpent   bool
	Realities     map[hash.Reality]bool
	Addresses     map[hash.Address]bool
}

func (options TransferOutputStorageFilters) Override(optionalFilters ...TransferOutputStorageFilter) TransferOutputStorageFilters {
	result := options
	for _, filter := range optionalFilters {
		filter(&result)
	}

	return result
}

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

func FilterRealities(realities ...hash.Reality) TransferOutputStorageFilter {
	return func(args *TransferOutputStorageFilters) {
		for _, reality := range realities {
			args.Realities[reality] = true
		}
	}
}

func FilterAddresses(addresses ...hash.Reality) TransferOutputStorageFilter {
	return func(args *TransferOutputStorageFilters) {
		for _, reality := range addresses {
			args.Addresses[reality] = true
		}
	}
}
