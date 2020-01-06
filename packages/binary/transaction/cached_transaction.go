package transaction

import (
	"github.com/iotaledger/hive.go/objectstorage"
)

type CachedTransaction struct {
	*objectstorage.CachedObject
}

func (cachedTransaction *CachedTransaction) Unwrap() *Transaction {
	if untypedTransaction := cachedTransaction.Get(); untypedTransaction == nil {
		return nil
	} else {
		if typeCastedTransaction := untypedTransaction.(*Transaction); typeCastedTransaction == nil || typeCastedTransaction.IsDeleted() {
			return nil
		} else {
			return typeCastedTransaction
		}
	}
}
