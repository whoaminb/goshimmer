package interfaces

import (
	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
)

type RealityStorage interface {
	LoadReality(realityId hash.Reality) (result Reality, err errors.IdentifiableError)
	StoreReality(reality Reality) (err errors.IdentifiableError)
}
