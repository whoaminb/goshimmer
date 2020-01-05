package tangle

import (
	"github.com/iotaledger/hive.go/events"
)

type tangleEvents struct {
	TransactionSolid    *events.Event
	TransactionAttached *events.Event
	Error               *events.Event
}
