package dkgapi

import (
	"encoding/hex"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"net/http"
)

func HandlerSignDigest(c echo.Context) error {
	var req SignDigestRequest

	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &SignDigestResponse{
			Err: err.Error(),
		})
	}
	return utils.ToJSON(c, http.StatusOK, SignDigestReq(&req))
}

type SignDigestRequest struct {
	AssemblyId *hashing.HashValue `json:"assembly_id"`
	Id         *hashing.HashValue `json:"id"`
	DataDigest *hashing.HashValue `json:"data_digest"`
}

type SignDigestResponse struct {
	SigShare string `json:"sig_share"`
	Err      string `json:"err"`
}

func SignDigestReq(req *SignDigestRequest) *SignDigestResponse {
	ks, ok, err := registry.GetDKShare(req.Id)
	if err != nil {
		return &SignDigestResponse{Err: err.Error()}
	}
	if !ok {
		return &SignDigestResponse{Err: "unknown key share"}
	}
	if !ks.Committed {
		return &SignDigestResponse{Err: "uncommitted key set"}
	}
	signature, err := ks.SignShare(req.DataDigest.Bytes())
	if err != nil {
		return &SignDigestResponse{Err: err.Error()}
	}
	return &SignDigestResponse{
		SigShare: hex.EncodeToString(signature),
	}
}
