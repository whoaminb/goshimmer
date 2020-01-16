package fpc

import (
	flag "github.com/spf13/pflag"
)

const (
	QUORUMSIZE     = "fpc.quorumSize"    // parameter.AddInt("FPC/QUORUMSIZE", 3, "size of Quorum for FPC")
	ROUND_INTERVAL = "fpc.roundInterval" //parameter.AddInt("FPC/ROUND_TIME", 5, "Time of Round")
	PRNG_ADDRESS   = "fpc.prngAddress"   //parameter.AddString("FPC/PRNG_ADDRESS", "127.0.0.1", "Centralized PRNG address")
	PRNG_PORT      = "fpc.prngPort"      //parameter.AddString("FPC/PRNG_PORT", "10000", "Centralized PRNG tcp port")
)

func init() {
	flag.Int(QUORUMSIZE, 3, "Size of the voting quorum (k)")
	flag.Int(ROUND_INTERVAL, 5, "FPC round interval [s]")
	flag.String(PRNG_ADDRESS, "127.0.0.1", "Centralized PRNG net address")
	flag.Int(PRNG_PORT, 10000, "Centralized PRNG tcp port")
}
