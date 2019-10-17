package transferoutput

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
)

type Reference struct {
	realityId    hash.Reality
	addressHash  hash.Address
	transferHash hash.Transfer
}

func NewReference(realityId hash.Reality, addressHash hash.Address, transferHash hash.Transfer) *Reference {
	return &Reference{
		realityId:    realityId,
		addressHash:  addressHash,
		transferHash: transferHash,
	}
}

func (reference *Reference) GetRealityId() hash.Reality {
	return reference.realityId
}

func (reference *Reference) GetAddressHash() hash.Address {
	return reference.addressHash
}

func (reference *Reference) GetTransferHash() hash.Transfer {
	return reference.transferHash
}
