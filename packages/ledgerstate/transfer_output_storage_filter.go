package ledgerstate

type TransferOutputStorageFilter func(*TransferOutputStorageFilters)

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
