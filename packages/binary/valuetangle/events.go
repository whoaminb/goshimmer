package valuetangle

import (
	"github.com/iotaledger/goshimmer/packages/binary/valuetangle/model"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
	"github.com/iotaledger/hive.go/events"
)

type Events struct {
	TransferAttached        *events.Event
	TransferSolid           *events.Event
	TransferMissing         *events.Event
	MissingTransferReceived *events.Event
	TransferRemoved         *events.Event
}

func newEvents() *Events {
	return &Events{
		TransferAttached:        events.NewEvent(cachedTransferEvent),
		TransferSolid:           events.NewEvent(cachedTransferEvent),
		TransferMissing:         events.NewEvent(transferIdEvent),
		MissingTransferReceived: events.NewEvent(transferIdEvent),
		TransferRemoved:         events.NewEvent(transferIdEvent),
	}
}

func transferIdEvent(handler interface{}, params ...interface{}) {
	missingTransactionId := params[0].(transfer.Id)

	handler.(func(transfer.Id))(missingTransactionId)
}

func cachedTransferEvent(handler interface{}, params ...interface{}) {
	cachedTransfer := params[0].(*model.CachedValueTransfer)
	cachedTransferMetadata := params[1].(*model.CachedTransferMetadata)

	cachedTransfer.RegisterConsumer()
	cachedTransferMetadata.RegisterConsumer()

	handler.(func(*model.CachedValueTransfer, *model.CachedTransferMetadata))(cachedTransfer, cachedTransferMetadata)
}
