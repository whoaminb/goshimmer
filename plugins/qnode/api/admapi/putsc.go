package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"net/http"
)

type PutSCDataRequest struct {
	ScId          string             `json:"sc_id"` // base58
	OwnerPubkey   *hashing.HashValue `json:"owner_pubkey"`
	Description   string             `json:"description"`
	NodeLocations []*registry.PortAddr
}

//----------------------------------------------------------
func HandlerPutSCData(c echo.Context) error {
	var req registry.SCData

	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{
			Error: err.Error(),
		})
	}
	if err := registry.SaveSCData(&req); err != nil {
		log.Errorf("failed to save SC data: %v", err)
		return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{Error: err.Error()})
	}
	log.Infof("SC data saved: scid = %s descr = '%s'", req.ScId.Short(), req.Description)

	return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{})
}
