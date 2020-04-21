package dkgapi

import (
	"encoding/hex"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"go.dedis.ch/kyber/v3/share"
	"net/http"
)

// The POST handler implements 'adm/getpubs' API
// Parameters(see GetPubsRequest struct):
//     assembly_id: assembly id: hex encoded 32 bytes of hash value
//     id:          distributed key set id, hex encoded 32 bytes of hash value
// API responds with:
// - pub_keys: list of public keys of corresponding nodes. This can be used to check individual signatures
//   of the particular node
// - pub_key_master: master public key, which allows to check BLS signatures by the quorum

func HandlerGetPubs(c echo.Context) error {
	var req GetPubsRequest

	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &GetPubsResponse{
			Err: err.Error(),
		})
	}
	return utils.ToJSON(c, http.StatusOK, GetPubsReq(&req))
}

type GetPubsRequest struct {
	AssemblyId *hashing.HashValue `json:"assembly_id"`
	Id         *hashing.HashValue `json:"id"`
}

type GetPubsResponse struct {
	PubKeys      []string `json:"pub_keys"`
	PubKeyMaster string   `json:"pub_key_master"`
	Err          string   `json:"err"`
}

func GetPubsReq(req *GetPubsRequest) *GetPubsResponse {
	ks, ok, err := registry.GetDKShare(req.Id)
	if err != nil {
		return &GetPubsResponse{Err: err.Error()}
	}
	if !ok {
		return &GetPubsResponse{Err: "unknown key share"}
	}
	if !ks.Committed {
		return &GetPubsResponse{Err: "uncommitted key set"}
	}
	pubkeys := make([]string, len(ks.PubKeys))
	for i, pk := range ks.PubKeys {
		pkb, err := pk.MarshalBinary()
		if err != nil {
			return &GetPubsResponse{Err: err.Error()}
		}
		pubkeys[i] = hex.EncodeToString(pkb)
	}
	pubShares := make([]*share.PubShare, len(ks.PubKeys))
	for i, v := range ks.PubKeys {
		pubShares[i] = &share.PubShare{
			I: i,
			V: v,
		}
	}
	pubPoly, err := share.RecoverPubPoly(ks.Suite.G2(), pubShares, int(ks.T), int(ks.N))
	if err != nil {
		panic(err)
	}
	pubKeyMaster := pubPoly.Commit()
	pkb, _ := pubKeyMaster.MarshalBinary()

	return &GetPubsResponse{
		PubKeys:      pubkeys,
		PubKeyMaster: hex.EncodeToString(pkb),
	}
}
