package qnode

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
)

const name = "Qnode"

var (
	PLUGIN  = node.NewPlugin(name, node.Enabled, config)
	PLUGINS = node.Plugins(PLUGIN)
	log     *logger.Logger
)

func config(_ *node.Plugin) {
	log = logger.NewLogger(name)
	log.Infof("Started Qnode plugin")
}
