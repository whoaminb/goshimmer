package api

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/admapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/dkgapi"
	"github.com/iotaledger/goshimmer/plugins/webapi"
	"github.com/iotaledger/hive.go/logger"
)

func InitEndpoints() {
	webapi.Server.POST("/adm/newdks", dkgapi.HandlerNewDks)
	webapi.Server.POST("/adm/aggregatedks", dkgapi.HandlerAggregateDks)
	webapi.Server.POST("/adm/commitdks", dkgapi.HandlerCommitDks)
	webapi.Server.POST("/adm/signdigest", dkgapi.HandlerSignDigest)
	webapi.Server.POST("/adm/getpubkeyinfo", dkgapi.HandlerGetKeyPubInfo)
	webapi.Server.POST("/adm/putscdata", admapi.HandlerPutSCData)
	webapi.Server.POST("/adm/getscdata", admapi.HandlerGetSCData)
	webapi.Server.GET("/adm/getsclist", admapi.HandlerGetSCList)

	logger.NewLogger("QnodeAPI").Infof("successfully added api endpoints")
}
