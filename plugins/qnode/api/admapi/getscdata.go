package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/transaction"
	"github.com/labstack/echo"
	"net/http"
)

type GetSCDataRequest struct {
	ScId *transaction.ScId `json:"sc_id"`
}

type GetSCDataResponse struct {
	registry.SCData
	Error string `json:"err"`
}

func HandlerGetSCData(c echo.Context) error {
	var req GetSCDataRequest

	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &GetSCDataResponse{
			Error: err.Error(),
		})
	}

	scdata, err := registry.GetSCData(req.ScId)
	if err != nil {
		return utils.ToJSON(c, http.StatusOK, &GetScListResponse{Error: err.Error()})
	}
	return utils.ToJSON(c, http.StatusOK, &GetSCDataResponse{SCData: *scdata})
}
