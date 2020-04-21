package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"net/http"
)

type GetScDataRequest struct {
	Id *hashing.HashValue `json:"id"`
}

type GetScDataResponse struct {
	registry.SCData
	Error string `json:"err"`
}

func HandlerGetSCData(c echo.Context) error {
	var req GetScDataRequest
	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &GetScDataResponse{Error: err.Error()})
	}
	// reads registry from DB into the cache
	if err := registry.RefreshSCDataCache(); err != nil {
		return utils.ToJSON(c, http.StatusOK, &GetScDataResponse{Error: err.Error()})
	}
	if req.Id == nil {
		return utils.ToJSON(c, http.StatusOK, &GetScDataResponse{Error: "wrong params"})
	}
	// retrieves SC data from the cache
	scData, ok := registry.GetScData(req.Id)
	if !ok {
		return utils.ToJSON(c, http.StatusOK, &GetScDataResponse{Error: "not found"})
	}
	return utils.ToJSON(c, http.StatusOK, &GetScDataResponse{SCData: *scData})
}
