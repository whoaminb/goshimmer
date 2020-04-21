package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/admapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/utils"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"net/http"
)

func PutSCData(addr string, port int, adata *registry.SCData) error {
	data, err := json.Marshal(adata)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("http://%s:%d/adm/scdata", addr, port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	var result utils.SimpleResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}
	if result.Error != "" {
		err = errors.New(result.Error)
	}
	return err
}

func GetSCdata(addr string, port int, schash *hashing.HashValue) (*registry.SCData, error) {
	req := admapi.GetScDataRequest{Id: schash}
	data, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s:%d/adm/getsc", addr, port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status %d", resp.StatusCode)
	}
	var dresp admapi.GetScDataResponse
	err = json.NewDecoder(resp.Body).Decode(&dresp)
	if err != nil {
		return nil, err
	}
	if dresp.Error != "" {
		return nil, errors.New(dresp.Error)
	}
	return &dresp.SCData, err
}

func GetSClist(url string) ([]*registry.SCData, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/adm/sclist", url))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status %d", resp.StatusCode)
	}
	var lresp admapi.GetScListResponse
	err = json.NewDecoder(resp.Body).Decode(&lresp)
	if err != nil {
		return nil, err
	}
	if lresp.Error != "" {
		return nil, errors.New(lresp.Error)
	}
	return lresp.SCList, nil
}
