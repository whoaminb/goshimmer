package transfer

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"
	"github.com/iotaledger/hive.go/objectstorage"
)

// region private utility methods //////////////////////////////////////////////////////////////////////////////////////

func GenerateOutputBookingStorageKey(realityId reality.Id, addressHash address.Address, spent bool, transferHash Id) (storageKey []byte) {
	storageKey = make([]byte, reality.IdLength+address.Length+1+IdLength)

	copy(storageKey[marshalTransferOutputBookingRealityIdStart:marshalTransferOutputBookingRealityIdEnd], realityId[:reality.IdLength])
	copy(storageKey[marshalTransferOutputBookingAddressHashStart:marshalTransferOutputBookingAddressHashEnd], addressHash[:address.Length])
	if spent {
		storageKey[marshalTransferOutputBookingSpentStart] = byte(SPENT)
	} else {
		storageKey[marshalTransferOutputBookingSpentStart] = byte(UNSPENT)
	}
	copy(storageKey[marshalTransferOutputBookingTransferHashStart:marshalTransferOutputBookingTransferHashEnd], transferHash[:IdLength])

	return
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

func OutputBookingFactory(key []byte) objectstorage.StorableObject {
	result := &OutputBooking{
		storageKey: make([]byte, len(key)),
	}
	copy(result.storageKey, key)

	return result
}
