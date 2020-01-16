package fpc

import (
	"net"
	"strconv"

	"github.com/iotaledger/goshimmer/packages/autopeering/peer/service"
	"github.com/iotaledger/goshimmer/packages/fpc"
	"github.com/iotaledger/goshimmer/packages/parameter"
	"github.com/iotaledger/goshimmer/plugins/autopeering/local"
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
	config   server.Config
)

func configureFPC() {
	log = logger.NewLogger(name)
	INSTANCE = fpc.New(network.GetKnownPeers, network.QueryNode, fpc.NewParameters())
	Events.VotingDone = events.NewEvent(votingDoneCaller)

	lPeer := local.GetInstance()

	port := strconv.Itoa(parameter.NodeConfig.GetInt(FPC_PORT))

	host, _, err := net.SplitHostPort(lPeer.Address())
	if err != nil {
		log.Fatalf("invalid peering address: %v", err)
	}
	err = lPeer.UpdateService(service.FPCKey, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		log.Fatalf("could not update services: %v", err)
	}

	config = server.Config{
		Address:        host,
		Port:           port,
		Log:            log,
		FPCInstance:    INSTANCE,
		ShutdownSignal: make(chan struct{}),
	}
}

func start(shutdownSignal <-chan struct{}) {
	defer log.Info("Stopping FPC Processor ... done")

	server.RunServer(config)

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
			config.ShutdownSignal <- struct{}{} //close FPC server
			log.Info("Stopping FPC Processor ...")
		}
	}

}
