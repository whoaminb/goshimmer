package model

import (
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/valuetransfer"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
)

type CachedValueTransfer struct {
	*transaction.CachedTransaction
}

func (cachedValueTransfer *CachedValueTransfer) Unwrap() (*valuetransfer.ValueTransfer, transfer.Id) {
	if untypedTransaction := cachedValueTransfer.Get(); untypedTransaction == nil {
		return nil, transfer.EmptyId
	} else {
		if typeCastedTransaction := untypedTransaction.(*transaction.Transaction); typeCastedTransaction == nil || typeCastedTransaction.IsDeleted() || typeCastedTransaction.GetPayload().GetType() != valuetransfer.Type {
			return nil, transfer.EmptyId
		} else {
			transactionId := typeCastedTransaction.GetId()

			return typeCastedTransaction.GetPayload().(*valuetransfer.ValueTransfer), transfer.NewId(transactionId[:])
		}
	}
}
