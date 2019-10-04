package ca

import (
	"github.com/iotaledger/goshimmer/packages/events"
	"github.com/iotaledger/goshimmer/packages/identity"
)

type HeartbeatManagerEvents struct {
	AddNeighbor    *events.Event
	RemoveNeighbor *events.Event
}

type NeighborManagerEvents struct {
	AddNeighbor      *events.Event
	RemoveNeighbor   *events.Event
	NeighborActive   *events.Event
	NeighborIdle     *events.Event
	ChainReset       *events.Event
	StatementMissing *events.Event
}

type StatementChainEvents struct {
	Reset *events.Event
}

func HashCaller(handler interface{}, params ...interface{}) {
	handler.(func([]byte))(params[0].([]byte))
}

func IdentityNeighborManagerCaller(handler interface{}, params ...interface{}) {
	handler.(func(*identity.Identity, *NeighborManager))(params[0].(*identity.Identity), params[1].(*NeighborManager))
}
