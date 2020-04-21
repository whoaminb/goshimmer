package dkgapi

import (
	"encoding/hex"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/labstack/echo"
	"go.dedis.ch/kyber/v3"
	"net/http"
)

// The POST handler implements 'adm/commitdks' API
// Parameters(see CommitDKSRequest struct):
//     assembly_id: assembly id: hex encoded 32 bytes of hash value
//     id:          distributed key set id, hex encoded 32 bytes of hash value
//     pub_shares: all public chares collected from all nodes
//
// Node does the following:
//    - finalizes all necessary information, which allows to sign data with the private key
//    in the way for it to be verifiable with public keys as per BLS threshold encryption
//    - saves the key into the registry of the nodes in persistent way
//
// After all adm/commitdks calls returns SUCCESS (no error), the dealer is sure, that ditributed
// key set with 'assembly_id' and 'id' was successfully distributed and persistently stored in respective nodes
//
// NOTE: the way the keys are distributed with 'newdks', 'aggregatedks' and 'commitdks' calls
// ensure that nobody, except at least 't' of nodes can create valid BLS signatures.
// Even dealer has not enough information to do that

func HandlerCommitDks(c echo.Context) error {
	var req CommitDKSRequest

	if err := c.Bind(&req); err != nil {
		return utils.ToJSON(c, http.StatusOK, &CommitDKSResponse{
			Err: err.Error(),
		})
	}
	return utils.ToJSON(c, http.StatusOK, CommitDKSReq(&req))
}

type CommitDKSRequest struct {
	Id        *hashing.HashValue `json:"id"`
	PubShares []string           `json:"pub_shares"`
}

type CommitDKSResponse struct {
	Address *hashing.HashValue `json:"address"`
	Err     string             `json:"err"`
}

func CommitDKSReq(req *CommitDKSRequest) *CommitDKSResponse {
	ks, ok, err := registry.GetDKShare(req.Id)
	if err != nil {
		return &CommitDKSResponse{Err: err.Error()}
	}
	if !ok {
		return &CommitDKSResponse{Err: "unknown key set"}
	}
	if ks.Committed {
		return &CommitDKSResponse{Err: "key set is already committed"}
	}
	if len(req.PubShares) != int(ks.N) {
		return &CommitDKSResponse{Err: "CommitDKSReq: wrong number of private shares"}
	}

	pubKeys := make([]kyber.Point, len(req.PubShares))
	for i, s := range req.PubShares {
		b, err := hex.DecodeString(s)
		if err != nil {
			return &CommitDKSResponse{Err: err.Error()}
		}
		p := ks.Suite.G2().Point()
		if err := p.UnmarshalBinary(b); err != nil {
			return &CommitDKSResponse{Err: err.Error()}
		}
		pubKeys[i] = p
	}
	err = registry.CommitDKShare(ks, pubKeys)
	if err != nil {
		return &CommitDKSResponse{Err: err.Error()}
	}
	registry.UncacheDKShare(req.Id)
	log.Infow("Created new key share",
		"address", ks.Address.String(),
		"N", ks.N,
		"T", ks.T,
		"Index", ks.Index,
	)
	return &CommitDKSResponse{
		Address: ks.Address,
	}
}
