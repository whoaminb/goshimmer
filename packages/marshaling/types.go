package marshaling

import "github.com/golang/protobuf/proto"

type ProtocolBufferTarget interface {
	ToProto() (result proto.Message)
	FromProto(proto proto.Message)
}
