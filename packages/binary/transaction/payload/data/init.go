package data

import (
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload"
)

func init() {
	payload.RegisterType(Type, GenericPayloadUnmarshalerFactory(Type))
}
