package interfaces

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
)

type TransferOutput interface {
	GetRealityId() hash.Reality
	GetAddressHash() hash.Address
	GetTransferHash() hash.Transfer
	GetConsumers() []hash.Transfer
	GetColoredBalances() map[hash.Color]ColoredBalance
}
