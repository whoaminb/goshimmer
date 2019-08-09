package marshal

import (
	"encoding/binary"
)

func Uint64(getter uint64Getter, setter uint64Setter) *uint64Field {
	return &uint64Field{
		getter: getter,
		setter: setter,
	}
}

type uint64Getter func(receiver interface{}) uint64
type uint64Setter func(receiver interface{}, val uint64)

type uint64Field struct {
	getter uint64Getter
	setter uint64Setter
}

func (uint64Field *uint64Field) Marshal(obj interface{}, data []byte) (result []byte) {
	result = make([]byte, 8)

	binary.BigEndian.PutUint64(result, obj.(uint64))

	return
}

func (uint64Field *uint64Field) Get(obj interface{}) interface{} {
	return uint64Field.getter(obj)
}
