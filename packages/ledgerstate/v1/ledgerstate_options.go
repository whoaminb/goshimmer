package v1

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/interfaces"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/storage/ram"
)

var DEFAULT_LEDGER_STATE_OPTIONS = &LedgerStateOptions{
	TransferOutputStorageFactory: ram.NewTransferOutputStorage,
	RealityStorageFactory:        ram.NewRealityStorage,
}

func OptionTransferOutputStorageFactory(factory interfaces.TransferOutputStorageFactory) LedgerStateOption {
	return func(args *LedgerStateOptions) {
		args.TransferOutputStorageFactory = factory
	}
}

func OptionStorageFactory(factory interfaces.RealityStorageFactory) LedgerStateOption {
	return func(args *LedgerStateOptions) {
		args.RealityStorageFactory = factory
	}
}

type LedgerStateOptions struct {
	TransferOutputStorageFactory interfaces.TransferOutputStorageFactory
	RealityStorageFactory        interfaces.RealityStorageFactory
}

func (options LedgerStateOptions) Override(optionalOptions ...LedgerStateOption) *LedgerStateOptions {
	result := &options
	for _, option := range optionalOptions {
		option(result)
	}

	return result
}

type LedgerStateOption func(*LedgerStateOptions)
