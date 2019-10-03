package ca

import (
	"github.com/iotaledger/goshimmer/packages/events"
)

type NeighborManagerEvents struct {
	ChainReset       *events.Event
	StatementMissing *events.Event
}
