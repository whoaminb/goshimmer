package fpc

import (
	flag "github.com/spf13/pflag"
)

const (
	QUORUMSIZE     = "fpc.quorumSize"
	ROUND_INTERVAL = "fpc.roundInterval"
	FPC_PORT       = "fpc.port"
	PRNG_ADDRESS   = "fpc.prngAddress"
	PRNG_PORT      = "fpc.prngPort"
)

func init() {
	flag.Int(QUORUMSIZE, 3, "Size of the voting quorum (k)")
	flag.Int(ROUND_INTERVAL, 5, "FPC round interval [s]")
	flag.Int(FPC_PORT, 10001, "FPC tcp port")
	flag.String(PRNG_ADDRESS, "127.0.0.1", "PRNG net address")
	flag.Int(PRNG_PORT, 10000, "PRNG tcp port")
}
