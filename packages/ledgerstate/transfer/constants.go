package transfer

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"
)

const (
	UNSPENT = SpentIndicator(0)
	SPENT   = SpentIndicator(1)

	marshalTransferOutputBookingRealityIdStart    = 0
	marshalTransferOutputBookingRealityIdEnd      = marshalTransferOutputBookingRealityIdStart + reality.IdLength
	marshalTransferOutputBookingAddressHashStart  = marshalTransferOutputBookingRealityIdEnd
	marshalTransferOutputBookingAddressHashEnd    = marshalTransferOutputBookingAddressHashStart + address.Length
	marshalTransferOutputBookingSpentStart        = marshalTransferOutputBookingAddressHashEnd
	marshalTransferOutputBookingSpentEnd          = marshalTransferOutputBookingSpentStart + 1
	marshalTransferOutputBookingTransferHashStart = marshalTransferOutputBookingSpentEnd
	marshalTransferOutputBookingTransferHashEnd   = marshalTransferOutputBookingTransferHashStart + IdLength
	marshalTransferOutputBookingTotalLength       = marshalTransferOutputBookingTransferHashEnd
)

type SpentIndicator byte
