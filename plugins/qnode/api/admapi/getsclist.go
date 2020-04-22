package admapi

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"net/http"
)

type GetScListResponse struct {
	SCDataList []*registry.SCData `json:"sc_data_list"`
	Error      string             `json:"err"`
}

func HandlerGetSCList(c echo.Context) error {
	sclist, err := registry.GetSCDataList(nil)
	if err != nil {
		return utils.ToJSON(c, http.StatusOK, &GetScListResponse{Error: err.Error()})
	}
	return utils.ToJSON(c, http.StatusOK, &GetScListResponse{SCDataList: sclist})
}
