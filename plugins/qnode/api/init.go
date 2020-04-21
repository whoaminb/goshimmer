package api

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/admapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/dkgapi"
	"github.com/iotaledger/goshimmer/plugins/webapi"
	"github.com/iotaledger/hive.go/logger"
)

func InitEndpoints() {
	//webapi.Server.GET("/kuku", TestRequest)

	webapi.Server.POST("/adm/newdks", dkgapi.HandlerNewDks)
	webapi.Server.POST("/adm/aggregatedks", dkgapi.HandlerAggregateDks)
	webapi.Server.POST("/adm/commitdks", dkgapi.HandlerCommitDks)
	webapi.Server.POST("/adm/signdigest", dkgapi.HandlerSignDigest)
	webapi.Server.POST("/adm/getpubs", dkgapi.HandlerGetPubs)
	webapi.Server.POST("/adm/newconfig", admapi.HandlerNewConfig)
	webapi.Server.POST("/adm/scdata", admapi.HandlerSCData)
	webapi.Server.POST("/adm/getsc", admapi.HandlerGetSCData)
	webapi.Server.GET("/adm/sclist", admapi.HandlerGetSCList)

	logger.NewLogger("QnodeAPI").Infof("successfully added api endpoints")
}

//
//func TestRequest(c echo.Context) error {
//	return c.String(http.StatusOK, "KUKU OK")
//}
