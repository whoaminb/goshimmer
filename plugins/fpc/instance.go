package fpc

import (
	"strconv"

	"github.com/iotaledger/goshimmer/packages/fpc"
	"github.com/iotaledger/goshimmer/packages/parameter"
	"github.com/iotaledger/goshimmer/plugins/fpc/network"
	"github.com/iotaledger/goshimmer/plugins/fpc/network/server"
	"github.com/iotaledger/goshimmer/plugins/fpc/prng/client"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
)

var (
	INSTANCE *fpc.Instance // INSTANCE points to the fpc instance (concrete type)
	log      *logger.Logger
	Events   fpcEvents // Events exposes fpc events
)

func configureFPC() {
	log = logger.NewLogger(name)
	INSTANCE = fpc.New(network.GetKnownPeers, network.QueryNode, fpc.NewParameters())
	Events.VotingDone = events.NewEvent(votingDoneCaller)
}

func start(shutdownSignal <-chan struct{}) {
	defer log.Info("Stopping FPC Processor ... done")

	var fpcServerShutdown chan struct{}
	server.RunServer(fpcServerShutdown, log, INSTANCE)

	ticker := client.NewTicker()
	ticker.Connect(parameter.NodeConfig.GetString(PRNG_ADDRESS) + ":" + strconv.Itoa(parameter.NodeConfig.GetInt(PRNG_PORT)))

	for {
		select {
		case newRandom := <-ticker.C:
			INSTANCE.Tick(newRandom.Index, newRandom.Value)
			//plugin.LogInfo(fmt.Sprintf("Round %v", newRandom.Index))
		case finalizedTxs := <-INSTANCE.FinalizedTxsChannel():
			// if len(finalizedTxs) == 0, an fpc round
			// ended with no new finalized transactions
			if len(finalizedTxs) > 0 {
				Events.VotingDone.Trigger(finalizedTxs)
			}
		case <-shutdownSignal:
			fpcServerShutdown <- struct{}{} //close FPC server
			log.Info("Stopping FPC Processor ...")
		}
	}

}
