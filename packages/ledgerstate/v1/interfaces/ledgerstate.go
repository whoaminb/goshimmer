package interfaces

import (
	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
)

type LedgerState interface {
	// Stores the new TransferOutput
	AddTransferOutput(transferOutput TransferOutput) LedgerState
	GetTransferOutput(transferOutputReference TransferOutputReference) TransferOutput
	ForEachTransferOutput(callback func(transferOutput TransferOutput), filter ...TransferOutputStorageFilter)

	BookTransfer(transfer Transfer) errors.IdentifiableError
	CreateReality(realityId hash.Reality) Reality
	GetReality(realityId hash.Reality) Reality
	ForEachReality(callback func(reality Reality), filter ...TransferOutputStorageFilter)

	// Merges the referenced realities into a new aggregated reality (if they are not conflicting). If the referenced
	// realities are conflicting or include a non-existing reality, this method returns nil. If the referenced realities
	// descend from one another, we return the "smallest common denominator", which means the deepest reality that
	// descends from all of the referenced realities.
	MergeRealities(realityIds ...hash.Reality) Reality
}
