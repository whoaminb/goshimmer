package social_consensus

import (
	"github.com/iotaledger/goshimmer/packages/events"
)

type NodeEvents struct {
	TransactionSolid *events.Event
}

func TransactionEventHandler(handler interface{}, params ...interface{}) {
	handler.(func(*Transaction))(params[0].(*Transaction))
}
