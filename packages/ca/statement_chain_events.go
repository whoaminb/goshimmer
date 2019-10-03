package ca

import (
	"github.com/iotaledger/goshimmer/packages/events"
)

type StatementChainEvents struct {
	Reset            *events.Event
	StatementMissing *events.Event
}

func HashCaller(handler interface{}, params ...interface{}) {
	handler.(func([]byte))(params[0].([]byte))
}
