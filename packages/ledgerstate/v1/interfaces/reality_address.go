package interfaces

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
)

type RealityAddress interface {
	// getters for the core properties
	GetHash() hash.Address
	GetRealityId() hash.Reality

	// additional utility methods of the entity
	GetReality() Reality
	GetBalances() map[hash.Color]uint64
	GetUnspentTransferOutputs() []TransferOutput
	GetTransferOutputs() map[hash.Transfer]TransferOutput
}
