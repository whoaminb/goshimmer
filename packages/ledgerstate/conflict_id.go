package ledgerstate

import (
	"unicode/utf8"

	"github.com/iotaledger/goshimmer/packages/stringify"
	"golang.org/x/crypto/blake2b"
)

type ConflictId [conflictSetIdLength]byte

func NewConflictSetId(conflictBytes ...interface{}) (result ConflictId) {
	switch len(conflictBytes) {
	case 2:
		transferHash, ok := conflictBytes[0].(TransferHash)
		if !ok {
			panic("expected first parameter of NewConflictSetId to be a TransferHash")
		}

		addressHash, ok := conflictBytes[0].(TransferHash)
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

func (conflictId *ConflictId) UnmarshalBinary(data []byte) error {
	copy(conflictId[:], data[:conflictSetIdLength])

	return nil
}

func (conflictId ConflictId) String() string {
	if utf8.Valid(conflictId[:]) {
		return string(conflictId[:])
	} else {
		return stringify.SliceOfBytes(conflictId[:])
	}
}

const conflictSetIdLength = 32
