package ledgerstate

const (
	UNSPENT = SpentIndicator(0)
	SPENT   = SpentIndicator(1)

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

type SpentIndicator byte

var (
	MAIN_REALITY_ID = NewRealityId("MAIN_REALITY")
)
