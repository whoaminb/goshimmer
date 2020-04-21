package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"net/http"
)

//----------------------------------------------------------
func HandlerSCData(c echo.Context) error {
	var req registry.SCData

	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{
			Error: err.Error(),
		})
	}

	if err := req.Save(); err != nil {
		log.Errorf("failed to save assembly data: %v", err)
		return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{Error: err.Error()})
	}
	log.Infof("SC data saved: id = %s descr = '%s'",
		req.Scid.Short(), req.Description)

	if err := registry.RefreshSCDataCache(); err != nil {
		return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{Error: err.Error()})
	}
	return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{})
}
