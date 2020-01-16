package fpc

import (
	"github.com/iotaledger/goshimmer/packages/shutdown"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/node"
)

const name = "FPC" // name of the plugin

var PLUGIN = node.NewPlugin(name, node.Enabled, configure, run)

func configure(plugin *node.Plugin) {
	configureFPC()
}

func run(plugin *node.Plugin) {
	if err := daemon.BackgroundWorker(name, start, shutdown.ShutdownPriorityFPC); err != nil {
		log.Errorf("Failed to start as daemon: %s", err)
	}
}
