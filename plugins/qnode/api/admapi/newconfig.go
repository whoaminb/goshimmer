package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"net/http"
)

func HandlerNewConfig(c echo.Context) error {
	var req registry.ConfigData

	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &utils.SimpleResponse{
			Error: err.Error(),
		})
	}
	//log.Debugf("HandlerNewConfig: %+v", req)
	return utils.ToJSON(c, http.StatusOK, NewConfigReq(&req))
}

type NewConfigResponse struct {
	ConfigId *hashing.HashValue `json:"config_id"`
	Err      string             `json:"err"`
}

func NewConfigReq(req *registry.ConfigData) *NewConfigResponse {
	req.ConfigId = registry.ConfigId(req)
	err := registry.ValidateConfig(req)
	if err != nil {
		return &NewConfigResponse{
			Err: err.Error(),
		}
	}
	err = registry.SaveConfig(req)
	if err != nil {
		return &NewConfigResponse{
			Err: err.Error(),
		}
	}
	log.Infow("Created new configuration", "scid", req.Scid, "config id", req.ConfigId)
	return &NewConfigResponse{
		ConfigId: req.ConfigId,
	}
}
