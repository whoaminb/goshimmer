package qnode

import (
	"github.com/iotaledger/goshimmer/plugins/config"
	"github.com/iotaledger/goshimmer/plugins/qnode/api"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/admapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/dkgapi"
	"github.com/iotaledger/goshimmer/plugins/webapi"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/node"
	"os"
)

const name = "Qnode"

var (
	PLUGIN  = node.NewPlugin(name, node.Enabled, configPlugin)
	PLUGINS = node.Plugins(PLUGIN)
	log     *logger.Logger
)

func configPlugin(_ *node.Plugin) {
	log = logger.NewLogger(name)
	admapi.InitLogger()
	dkgapi.InitLogger()

	log.Infof("Started Qnode plugin")

	cwd, _ := os.Getwd()
	log.Debugw("+++++ dbg",
		"current working dir", cwd,
		"bindAddress", config.Node.GetString(webapi.BIND_ADDRESS),
	)
	api.InitEndpoints()
}
