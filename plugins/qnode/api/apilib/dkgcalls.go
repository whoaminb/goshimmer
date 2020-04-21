package apilib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/dkgapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/pkg/errors"
	"net/http"
)

func callNewKey(addr string, port int, params dkgapi.NewDKSRequest) (*dkgapi.NewDKSResponse, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s:%d/adm/newdks", addr, port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	result := &dkgapi.NewDKSResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.Err == "" {
		return result, nil
	}
	return nil, errors.New(result.Err)
}

func callAggregate(addr string, port int, params dkgapi.AggregateDKSRequest) (*dkgapi.AggregateDKSResponse, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s:%d/adm/aggregatedks", addr, port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	result := &dkgapi.AggregateDKSResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.Err == "" {
		return result, nil
	}
	return nil, errors.New(result.Err)
}

func callCommit(addr string, port int, params dkgapi.CommitDKSRequest) (*hashing.HashValue, error) {
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s:%d/adm/commitdks", addr, port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	result := &dkgapi.CommitDKSResponse{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	if result.Err == "" {
		return result.Address, nil
	}
	return nil, errors.New(result.Err)
}
