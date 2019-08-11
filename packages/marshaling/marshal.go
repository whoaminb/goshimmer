package marshaling

import (
	"github.com/golang/protobuf/proto"
	"github.com/iotaledger/goshimmer/packages/errors"
)

func Marshal(serializable Serializable) (result []byte, err errors.IdentifiableError) {
	if marshaledData, marshalErr := proto.Marshal(serializable.ToProto()); marshalErr != nil {
		err = ErrMarshalFailed.Derive(marshalErr, "marshal failed")
	} else {
		result = marshaledData
	}

	return
}
