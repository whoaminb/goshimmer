package registry

import "github.com/iotaledger/hive.go/logger"

const modulename = "qnode/registry"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}
