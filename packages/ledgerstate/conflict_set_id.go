package ledgerstate

import (
	"unicode/utf8"

	"github.com/iotaledger/goshimmer/packages/stringify"
	"golang.org/x/crypto/blake2b"
)

type ConflictSetId [conflictSetIdLength]byte

func NewConflictSetId(conflictSetBytes ...interface{}) (result ConflictSetId) {
	switch len(conflictSetBytes) {
	case 2:
		transferHash, ok := conflictSetBytes[0].(TransferHash)
		if !ok {
			panic("expected first parameter of NewConflictSetId to be a TransferHash")
		}

		addressHash, ok := conflictSetBytes[0].(TransferHash)
		if !ok {
			panic("expected second parameter of NewConflictSetId to be a AddressHash")
		}

		fullConflictSetIdentifier := make([]byte, transferHashLength+addressHashLength)
		copy(fullConflictSetIdentifier, transferHash[:])
		copy(fullConflictSetIdentifier[transferHashLength:], addressHash[:])

		result = blake2b.Sum256(fullConflictSetIdentifier)
	case 1:
	default:
		panic("invalid parameter count when calling NewConflictSetId")
	}

	return
}

func (conflictSetId *ConflictSetId) UnmarshalBinary(data []byte) error {
	copy(conflictSetId[:], data[:conflictSetIdLength])

	return nil
}

func (conflictSetId ConflictSetId) String() string {
	if utf8.Valid(conflictSetId[:]) {
		return string(conflictSetId[:])
	} else {
		return stringify.SliceOfBytes(conflictSetId[:])
	}
}

const conflictSetIdLength = 32
