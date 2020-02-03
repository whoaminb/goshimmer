package transferoutput

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
)

type Id [transfer.IdLength + address.Length]byte

func NewId(addr address.Address, transferId transfer.Id) (result Id) {
	copy(result[:address.Length], addr[:])
	copy(result[address.Length:], transferId[:])

	return
}

func (id Id) GetAddress() (result address.Address) {
	copy(result[:], id[:address.Length])

	return
}

func (id Id) GetTransferId() (result transfer.Id) {
	copy(result[:], id[address.Length:])

	return
}

func (id *Id) UnmarshalBinary(data []byte) error {
	copy(id[:], data[:IdLength])

	return nil
}

func (id *Id) MarshalBinary() (data []byte, err error) {
	data = make([]byte, IdLength)
	copy(data[:], id[:])

	return
}

func (id Id) String() string {
	return id.GetAddress().String() + " :: " + id.GetTransferId().String()
}

const IdLength = address.Length + transfer.IdLength
