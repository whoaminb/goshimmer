package apilib

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/dkgapi"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/pkg/errors"
	"time"
)

func GenerateNewDistributedKeySet(nodes []*registry.PortAddr, n, t uint16) (*HashValue, error) {
	params := dkgapi.NewDKSRequest{
		N:  n,
		T:  t,
		Id: HashStrings(fmt.Sprintf("%v", time.Now())), // temporary id
	}
	if len(nodes) != int(params.N) {
		return nil, errors.New("len(nodes) != int(params.N)")
	}
	// generate new key shares
	// results in the matrix
	priSharesMatrix := make([][]string, params.N)
	for i, pa := range nodes {
		par := params
		par.Index = uint16(i)
		resp, err := callNewKey(pa.Addr, pa.Port, par)
		if err != nil {
			return nil, err
		}
		if len(resp.PriShares) != int(params.N) {
			return nil, errors.New("len(resp.PriShares) != int(params.N)")
		}
		priSharesMatrix[i] = resp.PriShares
	}

	// aggregate private shares
	pubShares := make([]string, params.N)
	priSharesCol := make([]string, params.N)
	for col, pa := range nodes {
		for row := range nodes {
			priSharesCol[row] = priSharesMatrix[row][col]
		}
		resp, err := callAggregate(pa.Addr, pa.Port, dkgapi.AggregateDKSRequest{
			Id:        params.Id,
			Index:     uint16(col),
			PriShares: priSharesCol,
		})
		if err != nil {
			return nil, err
		}
		pubShares[col] = resp.PubShare
	}

	// commit keys
	var accountRet *HashValue
	for _, pa := range nodes {
		account, err := callCommit(pa.Addr, pa.Port, dkgapi.CommitDKSRequest{
			Id:        params.Id,
			PubShares: pubShares,
		})
		if err != nil {
			return nil, err
		}
		if accountRet != nil && !accountRet.Equal(account) {
			return nil, errors.New("key commit return different hashes of master public key")
		}
		accountRet = account
	}
	return accountRet, nil
}
