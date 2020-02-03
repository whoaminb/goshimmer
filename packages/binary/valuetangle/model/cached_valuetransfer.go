package model

import (
	"github.com/iotaledger/goshimmer/packages/binary/tangle/model/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/tangle/model/transaction/payload/valuetransfer"
)

type CachedValueTransfer struct {
	*transaction.CachedTransaction
}

func (cachedValueTransfer *CachedValueTransfer) Unwrap() *ValueTransfer {
	if untypedTransaction := cachedValueTransfer.Get(); untypedTransaction == nil {
		return nil
	} else {
		if typeCastedTransaction := untypedTransaction.(*transaction.Transaction); typeCastedTransaction == nil || typeCastedTransaction.IsDeleted() || typeCastedTransaction.GetPayload().GetType() != valuetransfer.Type {
			return nil
		} else {
			return NewValueTransfer(typeCastedTransaction)
		}
	}
}
