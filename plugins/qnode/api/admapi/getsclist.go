package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"net/http"
)

type GetScListResponse struct {
	registry.SCList
	Error string `json:"err"`
}

func HandlerGetSCList(c echo.Context) error {
	sclist, ok := registry.GetSCDataList()
	if !ok {
		return utils.ToJSON(c, http.StatusOK, &GetScListResponse{Error: "not found"})
	}
	return utils.ToJSON(c, http.StatusOK, &GetScListResponse{SCList: sclist})
}
