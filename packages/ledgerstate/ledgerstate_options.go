package ledgerstate

var DEFAULT_LEDGER_STATE_OPTIONS = &LedgerStateOptions{
	TransferOutputStorageFactory: newTransferOutputStorageMemory,
	RealityStorageFactory:        newRealityStorageMemory,
}

func OptionTransferOutputStorageFactory(factory TransferOutputStorageFactory) LedgerStateOption {
	return func(args *LedgerStateOptions) {
		args.TransferOutputStorageFactory = factory
	}
}

func OptionStorageFactory(factory RealityStorageFactory) LedgerStateOption {
	return func(args *LedgerStateOptions) {
		args.RealityStorageFactory = factory
	}
}

type LedgerStateOptions struct {
	TransferOutputStorageFactory TransferOutputStorageFactory
	RealityStorageFactory        RealityStorageFactory
}

func (options LedgerStateOptions) Override(optionalOptions ...LedgerStateOption) *LedgerStateOptions {
	result := &options
	for _, option := range optionalOptions {
		option(result)
	}

	return result
}

type LedgerStateOption func(*LedgerStateOptions)
