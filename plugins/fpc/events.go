package fpc

import (
	"github.com/iotaledger/goshimmer/packages/fpc"
	"github.com/iotaledger/hive.go/events"
)

type fpcEvents struct {
	VotingDone *events.Event
}

func votingDoneCaller(handler interface{}, params ...interface{}) {
	handler.(func([]fpc.TxOpinion))(params[0].([]fpc.TxOpinion))
}
