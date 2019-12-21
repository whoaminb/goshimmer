package transaction

import (
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/data"
)

func init() {
	payload.SetGenericUnmarshalerFactory(data.GenericPayloadUnmarshalerFactory)
}
