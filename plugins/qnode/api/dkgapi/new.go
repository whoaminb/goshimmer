package dkgapi

import (
	"encoding/hex"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"github.com/labstack/echo"
	"net/http"
)

//----------------------------------------------------------
// The POST handler implements 'adm/newdks' API
// Parameters (see NewDKSRequest struct):
//     assembly_id: assembly id: hex encoded 32 bytes of hash value
//     id:          distributed key set id, hex encoded 32 bytes of hash value
//     n:           size of the assembly
//     t:           required quorum: normally t=floor( 2*n/3)+1
//	   index:       index of the node in the quorum
//
// This API must be called to n nodes with respective indices from 0 to n-1.
// Each node does the following:
// - generates random private key a0
// - generates random keys a1, ..., at
// - creates polynomial Pj(x) out with coefficients a0, a1, ... at of degree t,
//   where j is index of the node
//
// Response (see NewDKSResponse):
// - echoed values like id, assembly_id, index (probably not needed)
// - create - timestamp in unix milliseconds when created (probably not needed)
// - private shares PriShares: values of Pj(1), Pj(2).... Pj(n) EXCEPT diagonal value Pj(j)
//   where j == 'index'+1 of the current node
//
// After this call:
// - n nodes keeps in memory generated random polynomials (not saved yet)
// - caller nxn matrix of Private Shares (except the diagonal values).
//   j-th row of the matrix corresponds to private shares sent by the node j to the dealer (caller)
//
// In the next call 'adm/aggregatedks' dealer will be sending COLUMNS of the matrix to the same nodes.
// Note, that diagonal values never appear in public, so dealer is not able to reconstruct secret polynomials
//
// Next: see '/adm/aggregatedks' API

func HandlerNewDks(c echo.Context) error {
	var req NewDKSRequest

	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &NewDKSResponse{
			Err: err.Error(),
		})
	}
	return utils.ToJSON(c, http.StatusOK, NewDKSetReq(&req))
}

type NewDKSRequest struct {
	Id    *HashValue `json:"id"`
	N     uint16     `json:"n"`
	T     uint16     `json:"t"`
	Index uint16     `json:"index"` // 0 to N-1
}

type NewDKSResponse struct {
	PriShares []string `json:"pri_shares"`
	Err       string   `json:"err"`
}

func NewDKSetReq(req *NewDKSRequest) *NewDKSResponse {
	if err := tcrypto.ValidateDKSParams(req.T, req.N, req.Index); err != nil {
		return &NewDKSResponse{Err: err.Error()}
	}
	_, ok, err := registry.GetDKShare(req.Id)
	if err != nil {
		return &NewDKSResponse{Err: err.Error()}
	}
	if ok {
		return &NewDKSResponse{Err: "key set already exist"}
	}
	ks := tcrypto.NewRndDKShare(req.T, req.N, req.Index)
	registry.CacheDKShare(ks, req.Id)

	resp := NewDKSResponse{
		PriShares: make([]string, ks.N),
	}
	for i, s := range ks.PriShares {
		if uint16(i) != ks.Index {
			data, err := s.V.MarshalBinary()
			if err != nil {
				return &NewDKSResponse{Err: err.Error()}
			}
			resp.PriShares[i] = hex.EncodeToString(data)
		}
	}
	return &resp
}
