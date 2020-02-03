package model

import (
	"github.com/iotaledger/hive.go/objectstorage"
)

type CachedMissingTransfer struct {
	objectstorage.CachedObject
}

func (cachedObject *CachedMissingTransfer) Unwrap() *MissingTransfer {
	if untypedObject := cachedObject.Get(); untypedObject == nil {
		return nil
	} else {
		if typedObject := untypedObject.(*MissingTransfer); typedObject == nil || typedObject.IsDeleted() {
			return nil
		} else {
			return typedObject
		}
	}
}
