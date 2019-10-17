package interfaces

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
)

type TransferOutputReference interface {
	GetRealityId() hash.Reality
	GetAddressHash() hash.Address
	GetTransferHash() hash.Transfer
}
