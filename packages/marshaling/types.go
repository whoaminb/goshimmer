package marshaling

import "github.com/golang/protobuf/proto"

type Serializable interface {
	ToProto() (result proto.Message)
	FromProto(proto proto.Message)
}
