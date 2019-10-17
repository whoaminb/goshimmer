package interfaces

import (
	"github.com/iotaledger/goshimmer/packages/errors"
)

type TransferOutputStorage interface {
	LoadTransferOutput(transferOutputReference TransferOutputReference) (result TransferOutput, err errors.IdentifiableError)
	StoreTransferOutput(transferOutput TransferOutput) (err errors.IdentifiableError)
	ForEach(callback func(transferOutput TransferOutput), filters ...TransferOutputStorageFilter)
}
