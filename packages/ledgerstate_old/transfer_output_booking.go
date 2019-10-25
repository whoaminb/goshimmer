package ledgerstate

import (
	"github.com/iotaledger/goshimmer/packages/objectstorage"
)

// region struct + constructor + public api ////////////////////////////////////////////////////////////////////////////

type TransferOutputBooking struct {
	id           [marshalTransferOutputBookingTotalLength]byte
	realityId    RealityId
	addressHash  AddressHash
	spent        bool
	transferHash TransferHash
}

func newTransferOutputBooking(realityId RealityId, addressHash AddressHash, spent bool, transferHash TransferHash) (result *TransferOutputBooking) {
	result = &TransferOutputBooking{
		realityId:    realityId,
		addressHash:  addressHash,
		spent:        spent,
		transferHash: transferHash,
	}

	result.buildId()

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

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region support object storage ///////////////////////////////////////////////////////////////////////////////////////

func (booking *TransferOutputBooking) GetId() []byte {
	return booking.id[:]
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

func (booking *TransferOutputBooking) Marshal() ([]byte, error) {
	return []byte{}, nil
}

func (booking *TransferOutputBooking) Unmarshal(key []byte, serializedObject []byte) (objectstorage.StorableObject, error) {
	if len(key) < marshalTransferOutputBookingTotalLength {
		return nil, ErrUnmarshalFailed.Derive("unmarshal failed: the length of the key is to short")
	}

	result := &TransferOutputBooking{}

	copy(result.realityId[:], key[marshalTransferOutputBookingRealityIdStart:marshalTransferOutputBookingRealityIdEnd])
	copy(result.addressHash[:], key[marshalTransferOutputBookingAddressHashStart:marshalTransferOutputBookingAddressHashEnd])
	switch key[marshalTransferOutputBookingSpentStart] {
	case UNSPENT_SEPARATOR_BYTE:
		result.spent = false
	case SPENT_SEPARATOR_BYTE:
		result.spent = true
	default:
		return nil, ErrUnmarshalFailed.Derive("unmarshal failed: invalid spent separator in key")
	}
	copy(result.transferHash[:], key[marshalTransferOutputBookingTransferHashStart:marshalTransferOutputBookingTransferHashEnd])

	result.buildId()

	return result, nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region private utility methods //////////////////////////////////////////////////////////////////////////////////////

func (booking *TransferOutputBooking) buildId() {
	copy(booking.id[marshalTransferOutputBookingRealityIdStart:marshalTransferOutputBookingRealityIdEnd], booking.realityId[:realityIdLength])
	copy(booking.id[marshalTransferOutputBookingAddressHashStart:marshalTransferOutputBookingAddressHashEnd], booking.addressHash[:addressHashLength])
	if booking.spent {
		booking.id[marshalTransferOutputBookingSpentStart] = SPENT_SEPARATOR_BYTE
	} else {
		booking.id[marshalTransferOutputBookingSpentStart] = UNSPENT_SEPARATOR_BYTE
	}
	copy(booking.id[marshalTransferOutputBookingTransferHashStart:marshalTransferOutputBookingTransferHashEnd], booking.addressHash[:transferHashLength])
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
