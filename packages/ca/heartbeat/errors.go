package heartbeat

import (
	"github.com/iotaledger/goshimmer/packages/errors"
)

var (
	ErrSigningFailed = errors.Wrap(errors.New("failed to sign"), "failed to sign")
)
