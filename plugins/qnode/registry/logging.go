package registry

import "github.com/iotaledger/hive.go/logger"

const modulename = "qnode/registry"

var log *logger.Logger

func InitLogger() {
	log = logger.NewLogger(modulename)
}

func LogLoadedConfigs() {
	scDataMutex.Lock()
	defer scDataMutex.Unlock()

	log.Debugf("loaded %d assembly data record(s)", len(scDataCache))
	for aid, ad := range scDataCache {
		log.Debugw("assembly record", "aid", aid.Short(), "dscr", ad.Description)
	}
}
