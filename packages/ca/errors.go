package ca

import (
	"github.com/iotaledger/goshimmer/packages/errors"
)

var (
	ErrMalformedHeartbeat = errors.New("malformed heartbeat")
	ErrUnknownNeighbor    = errors.New("unknown neighbor")
	ErrTooManyNeighbors   = errors.New("too many neighbors")
)
