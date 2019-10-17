package interfaces

import (
	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
)

type Reality interface {
	// getters for the core properties
	GetId() hash.Reality
	GetParentRealityIds() []hash.Reality

	// additional utility methods of the entity
	GetParentRealities() map[hash.Reality]Reality
	GetAncestorRealities() map[hash.Reality]Reality
	DescendsFromReality(realityId hash.Reality) bool
	GetAddress(addressHash hash.Address) RealityAddress
	BookTransfer(transfer Transfer) errors.IdentifiableError
}
