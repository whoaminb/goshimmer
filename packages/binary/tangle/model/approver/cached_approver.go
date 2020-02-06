package approver

import (
	"github.com/iotaledger/hive.go/objectstorage"
)

type CachedApprover struct {
	objectstorage.CachedObject
}

func (cachedApprover *CachedApprover) Unwrap() *Approver {
	if untypedObject := cachedApprover.Get(); untypedObject == nil {
		return nil
	} else {
		if typedObject := untypedObject.(*Approver); typedObject == nil || typedObject.IsDeleted() {
			return nil
		} else {
			return typedObject
		}
	}
}
