package ca

import (
	"github.com/iotaledger/goshimmer/packages/events"
	"github.com/iotaledger/goshimmer/packages/identity"
)

type HeartbeatManagerEvents struct {
	AddNeighbor    *events.Event
	RemoveNeighbor *events.Event
}

func IdentityNeighborManagerCaller(handler interface{}, params ...interface{}) {
	handler.(func(*identity.Identity, *NeighborManager))(params[0].(*identity.Identity), params[1].(*NeighborManager))
}
