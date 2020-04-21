package registry

import "fmt"

type PortAddr struct {
	Port int    `json:"port"`
	Addr string `json:"addr"`
}

func (oa *PortAddr) AdjustedIP() (string, int) {
	if oa.Addr == "localhost" {
		return "127.0.0.1", oa.Port
	}
	return oa.Addr, oa.Port
}

func (oa *PortAddr) String() string {
	return fmt.Sprintf("%s:%d", oa.Addr, oa.Port)
}
