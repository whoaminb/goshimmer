package fcob

import (
	"github.com/iotaledger/goshimmer/plugins/fpc"
	"github.com/iotaledger/goshimmer/plugins/tangle"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
)

const name = "FCOB" // name of the plugin

// PLUGIN is the exposed FCoB plugin
var PLUGIN = node.NewPlugin(name, node.Enabled, configure, run)

var (
	api            tangleStore
	runFCOB        *events.Closure
	updateTxsVoted *events.Closure
	log            *logger.Logger
)

func configure(plugin *node.Plugin) {
	log = logger.NewLogger(name)
	api = tangleStore{}
	runFCOB = configureFCOB(log, api, fpc.INSTANCE)
	updateTxsVoted = configureUpdateTxsVoted(plugin, api)
}

func run(plugin *node.Plugin) {
	// subscribe to a new Tx solid event
	// and start an instance of the FCoB protocol
	tangle.Events.TransactionSolid.Attach(runFCOB)

	// subscribe to a new VotingDone event
	// and update the related txs opinion
	fpc.Events.VotingDone.Attach(updateTxsVoted)
}
