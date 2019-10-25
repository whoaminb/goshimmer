package ledgerstate

var (
	MAIN_REALITY_ID = [realityIdLength]byte{}
)

const (
	UNSPENT_SEPARATOR_BYTE = byte(0)
	SPENT_SEPARATOR_BYTE   = byte(1)

	transferHashLength = 32
	addressHashLength  = 32
	realityIdLength    = 32

	marshalTransferOutputBookingRealityIdStart    = 0
	marshalTransferOutputBookingRealityIdEnd      = marshalTransferOutputBookingRealityIdStart + realityIdLength
	marshalTransferOutputBookingAddressHashStart  = marshalTransferOutputBookingRealityIdEnd
	marshalTransferOutputBookingAddressHashEnd    = marshalTransferOutputBookingAddressHashStart + addressHashLength
	marshalTransferOutputBookingSpentStart        = marshalTransferOutputBookingAddressHashEnd
	marshalTransferOutputBookingSpentEnd          = marshalTransferOutputBookingSpentStart + 1
	marshalTransferOutputBookingTransferHashStart = marshalTransferOutputBookingSpentEnd
	marshalTransferOutputBookingTransferHashEnd   = marshalTransferOutputBookingTransferHashStart + transferHashLength
	marshalTransferOutputBookingTotalLength       = marshalTransferOutputBookingTransferHashEnd
)
