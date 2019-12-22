package conflict

import (
	"unicode/utf8"

	"github.com/iotaledger/goshimmer/packages/binary/transfer"

	"github.com/iotaledger/goshimmer/packages/binary/address"

	"github.com/iotaledger/goshimmer/packages/stringify"
	"golang.org/x/crypto/blake2b"
)

type Id [IdLength]byte

func NewId(conflictBytes ...interface{}) (result Id) {
	switch len(conflictBytes) {
	case 2:
		transferHash, ok := conflictBytes[0].(transfer.Hash)
		if !ok {
			panic("expected first parameter of NewId to be a TransferHash")
		}

		addressHash, ok := conflictBytes[0].(transfer.Hash)
		if !ok {
			panic("expected second parameter of NewId to be a AddressHash")
		}

		fullConflictSetIdentifier := make([]byte, transfer.HashLength+address.Length)
		copy(fullConflictSetIdentifier, transferHash[:])
		copy(fullConflictSetIdentifier[transfer.HashLength:], addressHash[:])

		result = blake2b.Sum256(fullConflictSetIdentifier)
	case 1:
	default:
		panic("invalid parameter count when calling NewId")
	}

	return
}

func (conflictId *Id) UnmarshalBinary(data []byte) error {
	copy(conflictId[:], data[:IdLength])

	return nil
}

func (conflictId Id) String() string {
	if utf8.Valid(conflictId[:]) {
		return string(conflictId[:])
	} else {
		return stringify.SliceOfBytes(conflictId[:])
	}
}

const IdLength = 32
