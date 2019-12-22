package ledgerstate

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
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
	marshalTransferOutputBookingTransferHashEnd   = marshalTransferOutputBookingTransferHashStart + transfer.HashLength
	marshalTransferOutputBookingTotalLength       = marshalTransferOutputBookingTransferHashEnd
)

type SpentIndicator byte

var (
	MAIN_REALITY_ID = reality.NewId("MAIN_REALITY")
)
