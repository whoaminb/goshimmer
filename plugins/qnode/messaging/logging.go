package messaging

import "github.com/iotaledger/hive.go/logger"

const modulename = "qnode/messaging"

var log *logger.Logger

func initLogger() {
	log = logger.NewLogger(modulename)
}
