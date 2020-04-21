package apilib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/admapi"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"net/http"
)

func NewConfiguration(addr string, port int, cdata *registry.ConfigData) (*hashing.HashValue, error) {
	data, err := json.Marshal(cdata)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("http://%s:%d/adm/newconfig", addr, port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	var result admapi.NewConfigResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	err = nil
	if result.Err != "" {
		return nil, errors.New(result.Err)
	}
	return result.ConfigId, nil
}
