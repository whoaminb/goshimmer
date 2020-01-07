package valuetangle

import (
	"github.com/iotaledger/hive.go/events"
)

type Events struct {
	TransferAttached        *events.Event
	TransferSolid           *events.Event
	TransferMissing         *events.Event
	MissingTransferReceived *events.Event
	TransferRemoved         *events.Event
}
