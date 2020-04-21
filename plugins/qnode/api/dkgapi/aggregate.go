package dkgapi

import (
	"encoding/hex"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"go.dedis.ch/kyber/v3"
	"net/http"
)

//----------------------------------------------------------
// The POST handler implements 'adm/aggregatedks' API
// Parameters(see AggregateDKSRequest struct):
//     assembly_id: assembly id: hex encoded 32 bytes of hash value
//     id:          distributed key set id, hex encoded 32 bytes of hash value
//     index:       index of the node in the assembly
//         (node knows it from the previous adm/newdks call, ths parameter is for control only)
//     pri_shares: values P1(j), P2(j), ...., Pn(j) EXCEPT Pj(j), the diagonal
//        where j is index+1 of the called node.
//
// Node does the following:
//  - it sums up all received pri_shares and own diagonal privat share which it kept for itself
//    The result is private share with number j  of the master secret polynomial,
//    which is not know by anybody, only by this node
// - It calculates public share from the private one
//
// Node's response (see AggregateDKSResponse struct)
// - Index is just for control
// - PubShare, calculated from private share
//
// After response from all nodes, dealer has all public information and nodes have all private informations.
// Key set is not saved yet!
// Next: see API call 'adm/commitdks'

func HandlerAggregateDks(c echo.Context) error {
	var req AggregateDKSRequest

	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &AggregateDKSResponse{
			Err: err.Error(),
		})
	}
	return utils.ToJSON(c, http.StatusOK, AggregateDKSReq(&req))
}

type AggregateDKSRequest struct {
	Id        *hashing.HashValue `json:"id"`
	Index     uint16             `json:"index"` // 0 to N-1
	PriShares []string           `json:"pri_shares"`
}

type AggregateDKSResponse struct {
	PubShare string `json:"pub_share"`
	Err      string `json:"err"`
}

func AggregateDKSReq(req *AggregateDKSRequest) *AggregateDKSResponse {
	ks, ok, err := registry.GetDKShare(req.Id)
	if err != nil {
		return &AggregateDKSResponse{Err: err.Error()}
	}
	if !ok {
		return &AggregateDKSResponse{Err: "unknown key set"}
	}
	if ks.Aggregated {
		return &AggregateDKSResponse{Err: "key set already aggregated"}
	}
	if len(req.PriShares) != int(ks.N) {
		return &AggregateDKSResponse{Err: "wrong number of private shares"}
	}
	if req.Index != ks.Index {
		return &AggregateDKSResponse{Err: "wrong index"}
	}
	// aggregate secret shares
	priShares := make([]kyber.Scalar, ks.N)
	for i, pks := range req.PriShares {
		if uint16(i) == ks.Index {
			continue
		}
		pkb, err := hex.DecodeString(pks)
		if err != nil {
			return &AggregateDKSResponse{Err: fmt.Sprintf("decode error: %v", err)}
		}
		priShares[i] = ks.Suite.G2().Scalar()
		if err := priShares[i].UnmarshalBinary(pkb); err != nil {
			return &AggregateDKSResponse{Err: fmt.Sprintf("unmarshal error: %v", err)}
		}
	}
	if err := ks.AggregateDKS(priShares); err != nil {
		return &AggregateDKSResponse{Err: fmt.Sprintf("aggregate error 1: %v", err)}
	}
	pkb, err := ks.PubKeyOwn.MarshalBinary()
	if err != nil {
		return &AggregateDKSResponse{Err: fmt.Sprintf("marshal error 2: %v", err)}
	}
	return &AggregateDKSResponse{
		PubShare: hex.EncodeToString(pkb),
	}
}
