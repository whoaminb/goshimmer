package ca

import (
	"github.com/iotaledger/goshimmer/packages/ca/heartbeat"
)

type HeartbeatChain struct {
	missingHeartbeats map[string]bool
	heartbeats        map[string]*heartbeat.Heartbeat
}

func NewHeartbeatChain() *HeartbeatChain {
	return &HeartbeatChain{
		heartbeats: make(map[string]*heartbeat.Heartbeat),
	}
}
