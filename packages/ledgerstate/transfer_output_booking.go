package ledgerstate

import (
	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/stringify"
	"github.com/iotaledger/hive.go/objectstorage"
)

// region struct + constructor + public api ////////////////////////////////////////////////////////////////////////////

type TransferOutputBooking struct {
	objectstorage.StorableObjectFlags

	realityId    RealityId
	addressHash  AddressHash
	spent        bool
	transferHash TransferHash

	storageKey []byte
}

func newTransferOutputBooking(realityId RealityId, addressHash AddressHash, spent bool, transferHash TransferHash) (result *TransferOutputBooking) {
	result = &TransferOutputBooking{
		realityId:    realityId,
		addressHash:  addressHash,
		spent:        spent,
		transferHash: transferHash,

		storageKey: generateTransferOutputBookingStorageKey(realityId, addressHash, spent, transferHash),
	}

	return
}

func (booking *TransferOutputBooking) GetRealityId() RealityId {
	return booking.realityId
}

func (booking *TransferOutputBooking) GetAddressHash() AddressHash {
	return booking.addressHash
}

func (booking *TransferOutputBooking) IsSpent() bool {
	return booking.spent
}

func (booking *TransferOutputBooking) GetTransferHash() TransferHash {
	return booking.transferHash
}

func (booking *TransferOutputBooking) String() string {
	return stringify.Struct("TransferOutputBooking",
		stringify.StructField("realityId", booking.realityId),
		stringify.StructField("addressHash", booking.addressHash),
		stringify.StructField("spent", booking.spent),
		stringify.StructField("transferHash", booking.transferHash),
	)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region support object storage ///////////////////////////////////////////////////////////////////////////////////////

func (booking *TransferOutputBooking) GetStorageKey() []byte {
	return booking.storageKey
}

func (booking *TransferOutputBooking) Update(other objectstorage.StorableObject) {
	if otherBooking, ok := other.(*TransferOutputBooking); !ok {
		panic("Update method expects a *TransferOutputBooking")
	} else {
		booking.realityId = otherBooking.realityId
		booking.addressHash = otherBooking.addressHash
		booking.spent = otherBooking.spent
		booking.transferHash = otherBooking.transferHash
	}
}

func (booking *TransferOutputBooking) MarshalBinary() ([]byte, error) {
	return []byte{}, nil
}

func (booking *TransferOutputBooking) UnmarshalBinary(data []byte) error {
	if len(booking.storageKey) < marshalTransferOutputBookingTotalLength {
		return errors.New("unmarshal failed: the length of the key is to short")
	}

	if err := booking.realityId.UnmarshalBinary(booking.storageKey[marshalTransferOutputBookingRealityIdStart:marshalTransferOutputBookingRealityIdEnd]); err != nil {
		return err
	}

	if err := booking.addressHash.UnmarshalBinary(booking.storageKey[marshalTransferOutputBookingAddressHashStart:marshalTransferOutputBookingAddressHashEnd]); err != nil {
		return err
	}

	switch SpentIndicator(booking.storageKey[marshalTransferOutputBookingSpentStart]) {
	case UNSPENT:
		booking.spent = false
	case SPENT:
		booking.spent = true
	default:
		return errors.New("unmarshal failed: invalid spent separator in key")
	}

	if err := booking.transferHash.UnmarshalBinary(booking.storageKey[marshalTransferOutputBookingTransferHashStart:marshalTransferOutputBookingTransferHashEnd]); err != nil {
		return err
	}

	return nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region private utility methods //////////////////////////////////////////////////////////////////////////////////////

func generateTransferOutputBookingStorageKey(realityId RealityId, addressHash AddressHash, spent bool, transferHash TransferHash) (storageKey []byte) {
	storageKey = make([]byte, realityIdLength+addressHashLength+1+transferHashLength)

	copy(storageKey[marshalTransferOutputBookingRealityIdStart:marshalTransferOutputBookingRealityIdEnd], realityId[:realityIdLength])
	copy(storageKey[marshalTransferOutputBookingAddressHashStart:marshalTransferOutputBookingAddressHashEnd], addressHash[:addressHashLength])
	if spent {
		storageKey[marshalTransferOutputBookingSpentStart] = byte(SPENT)
	} else {
		storageKey[marshalTransferOutputBookingSpentStart] = byte(UNSPENT)
	}
	copy(storageKey[marshalTransferOutputBookingTransferHashStart:marshalTransferOutputBookingTransferHashEnd], transferHash[:transferHashLength])

	return
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
