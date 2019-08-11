package marshaling

import "github.com/iotaledger/goshimmer/packages/errors"

var (
	ErrUnmarshalFailed = errors.Wrap(errors.New("unmarshal failed"), "unmarshal failed")
	ErrMarshalFailed   = errors.Wrap(errors.New("marshal failed"), "marshal failed")
)
