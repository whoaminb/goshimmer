package main

import (
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/api/apilib"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"io/ioutil"
	"os"
)

type ioParams struct {
	Hosts     []*registry.PortAddr `json:"hosts"`
	N         uint16               `json:"n"`
	T         uint16               `json:"t"`
	NumKeys   uint16               `json:"num_keys"`
	Addresses []*hashing.HashValue `json:"addresses"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage newdks <input file path>\n")
		os.Exit(1)
	}
	fname := os.Args[1]
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	params := ioParams{}
	err = json.Unmarshal(data, &params)
	if err != nil {
		panic(err)
	}
	if len(params.Hosts) != int(params.N) || params.N < params.T || params.N < 4 {
		panic("wrong assembly size parameters or number rof hosts")
	}

	params.Addresses = make([]*hashing.HashValue, 0, params.NumKeys)
	for i := 0; i < int(params.NumKeys); i++ {
		addr, err := apilib.GenerateNewDistributedKeySet(params.Hosts, params.N, params.T)
		params.Addresses = append(params.Addresses, addr)
		if err == nil {
			fmt.Printf("generated new keys. Address = %s\n", addr.String())
		} else {
			fmt.Printf("error: %v\n", err)
		}
	}
	data, err = json.MarshalIndent(&params, "", " ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	err = ioutil.WriteFile(fname+".resp.json", data, 0644)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
}
