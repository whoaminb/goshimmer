package dkgapi

import (
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/address"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"github.com/mr-tron/base58"
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
	Address    string             `json:"address"`
	DataDigest *hashing.HashValue `json:"data_digest"`
}

type SignDigestResponse struct {
	SigShare string `json:"sig_share"`
	Err      string `json:"err"`
}

func SignDigestReq(req *SignDigestRequest) *SignDigestResponse {
	addr, err := address.FromBase58(req.Address)
	if err != nil {
		return &SignDigestResponse{Err: err.Error()}
	}
	if addr.Version() != address.VERSION_BLS {
		return &SignDigestResponse{Err: "expected BLS address"}
	}
	ks, ok, err := registry.GetDKShare(&addr)
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
		SigShare: base58.Encode(signature),
	}
}
