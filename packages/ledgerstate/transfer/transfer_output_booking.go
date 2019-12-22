package transfer

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"
	"github.com/iotaledger/goshimmer/packages/stringify"
	"github.com/iotaledger/hive.go/objectstorage"
)

// region struct + constructor + public api ////////////////////////////////////////////////////////////////////////////

type OutputBooking struct {
	objectstorage.StorableObjectFlags

	realityId    reality.Id
	addressHash  address.Address
	spent        bool
	transferHash Hash

	storageKey []byte
}

func NewTransferOutputBooking(realityId reality.Id, addressHash address.Address, spent bool, transferHash Hash) (result *OutputBooking) {
	result = &OutputBooking{
		realityId:    realityId,
		addressHash:  addressHash,
		spent:        spent,
		transferHash: transferHash,

		storageKey: GenerateOutputBookingStorageKey(realityId, addressHash, spent, transferHash),
	}

	return
}

func (booking *OutputBooking) GetRealityId() reality.Id {
	return booking.realityId
}

func (booking *OutputBooking) GetAddressHash() address.Address {
	return booking.addressHash
}

func (booking *OutputBooking) IsSpent() bool {
	return booking.spent
}

func (booking *OutputBooking) GetTransferHash() Hash {
	return booking.transferHash
}

func (booking *OutputBooking) String() string {
	return stringify.Struct("OutputBooking",
		stringify.StructField("realityId", booking.realityId),
		stringify.StructField("addressHash", booking.addressHash),
		stringify.StructField("spent", booking.spent),
		stringify.StructField("transferHash", booking.transferHash),
	)
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region support object storage ///////////////////////////////////////////////////////////////////////////////////////

func (booking *OutputBooking) GetStorageKey() []byte {
	return booking.storageKey
}

func (booking *OutputBooking) Update(other objectstorage.StorableObject) {
	if otherBooking, ok := other.(*OutputBooking); !ok {
		panic("Update method expects a *OutputBooking")
	} else {
		booking.realityId = otherBooking.realityId
		booking.addressHash = otherBooking.addressHash
		booking.spent = otherBooking.spent
		booking.transferHash = otherBooking.transferHash
	}
}

func (booking *OutputBooking) MarshalBinary() ([]byte, error) {
	return []byte{}, nil
}

func (booking *OutputBooking) UnmarshalBinary(data []byte) error {
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
