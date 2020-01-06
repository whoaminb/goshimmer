package approvers

import (
	"github.com/iotaledger/hive.go/objectstorage"
)

type CachedApprovers struct {
	*objectstorage.CachedObject
}

func (cachedObject *CachedApprovers) Unwrap() *Approvers {
	if untypedObject := cachedObject.Get(); untypedObject == nil {
		return nil
	} else {
		if typedObject := untypedObject.(*Approvers); typedObject == nil || typedObject.IsDeleted() {
			return nil
		} else {
			return typedObject
		}
	}
}
