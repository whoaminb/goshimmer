package marshaling

import (
	"github.com/golang/protobuf/proto"
	"github.com/iotaledger/goshimmer/packages/errors"
)

func Unmarshal(serializable Serializable, protobuf proto.Message, data []byte) (err errors.IdentifiableError) {
	if unmarshalError := proto.Unmarshal(data, protobuf); unmarshalError != nil {
		err = ErrUnmarshalFailed.Derive(unmarshalError, "unmarshal failed")
	} else {
		serializable.FromProto(protobuf)
	}

	return
}
