package marshaling

import (
	"github.com/golang/protobuf/proto"
	"github.com/iotaledger/goshimmer/packages/errors"
)

// Unmarshals the given data into the target using the given protobuf type.
func Unmarshal(target ProtocolBufferTarget, data []byte, messageType proto.Message) (err errors.IdentifiableError) {
	if unmarshalError := proto.Unmarshal(data, messageType); unmarshalError != nil {
		err = ErrUnmarshalFailed.Derive(unmarshalError, "unmarshal failed")
	} else {
		target.FromProto(messageType)
	}

	return
}
