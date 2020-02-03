package model

import (
	"github.com/iotaledger/hive.go/objectstorage"
)

type CachedConsumers struct {
	objectstorage.CachedObject
}

func (cachedObject *CachedConsumers) Unwrap() *Consumers {
	if untypedObject := cachedObject.Get(); untypedObject == nil {
		return nil
	} else {
		if typedObject := untypedObject.(*Consumers); typedObject == nil || typedObject.IsDeleted() {
			return nil
		} else {
			return typedObject
		}
	}
}
