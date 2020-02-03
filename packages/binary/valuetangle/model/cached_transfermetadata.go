package model

import (
	"github.com/iotaledger/hive.go/objectstorage"
)

type CachedTransferMetadata struct {
	objectstorage.CachedObject
}

func (cachedObject *CachedTransferMetadata) Unwrap() *TransferMetadata {
	if untypedObject := cachedObject.Get(); untypedObject == nil {
		return nil
	} else {
		if typedObject := untypedObject.(*TransferMetadata); typedObject == nil || typedObject.IsDeleted() {
			return nil
		} else {
			return typedObject
		}
	}
}
