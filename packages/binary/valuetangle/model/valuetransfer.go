package model

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/valuetransfer"
	"github.com/iotaledger/goshimmer/packages/binary/types"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
)

type ValueTransfer struct {
	*transaction.Transaction
	*valuetransfer.ValueTransfer

	inputs      map[transfer.Id]map[address.Address]types.Empty
	inputsMutex sync.RWMutex
}

func NewValueTransfer(transaction *transaction.Transaction) *ValueTransfer {
	return &ValueTransfer{
		Transaction:   transaction,
		ValueTransfer: transaction.GetPayload().(*valuetransfer.ValueTransfer),
	}
}

func (valueTransfer *ValueTransfer) GetId() transfer.Id {
	transactionId := valueTransfer.Transaction.GetId()

	return transfer.NewId(transactionId[:])
}

func (valueTransfer *ValueTransfer) GetInputs() (result map[transfer.Id]map[address.Address]types.Empty) {
	valueTransfer.inputsMutex.RLock()
	if valueTransfer.inputs == nil {
		valueTransfer.inputsMutex.RUnlock()

		valueTransfer.inputsMutex.Lock()
		if valueTransfer.inputs == nil {
			result = make(map[transfer.Id]map[address.Address]types.Empty)

			for _, transferOutputReference := range valueTransfer.ValueTransfer.GetInputs() {
				addressMap, addressMapExists := result[transferOutputReference.GetTransferHash()]
				if !addressMapExists {
					addressMap = make(map[address.Address]types.Empty)

					result[transferOutputReference.GetTransferHash()] = addressMap
				}

				addressMap[transferOutputReference.GetAddress()] = types.Void
			}

			valueTransfer.inputs = result
		} else {
			result = valueTransfer.inputs
		}
		valueTransfer.inputsMutex.Unlock()
	} else {
		result = valueTransfer.inputs
	}

	return
}
