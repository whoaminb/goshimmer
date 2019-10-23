package test

import (
	"encoding/binary"

	"github.com/iotaledger/goshimmer/packages/objectstorage"
)

type TestObject struct {
	id    []byte
	value uint32
}

func NewTestObject(id string, value uint32) *TestObject {
	return &TestObject{
		id:    []byte(id),
		value: value,
	}
}

func (testObject *TestObject) GetId() []byte {
	return testObject.id
}

func (testObject *TestObject) Marshal() ([]byte, error) {
	result := make([]byte, 4)

	binary.LittleEndian.PutUint32(result, testObject.value)

	return result, nil
}

func (testObject *TestObject) Update(object objectstorage.StorableObject) {
	if obj, ok := object.(*TestObject); !ok {
		panic("invalid object passed to testObject.Update()")
	} else {
		testObject.value = obj.value
	}
}

func (testObject *TestObject) Unmarshal(key []byte, data []byte) (object objectstorage.StorableObject, e error) {
	return &TestObject{
		value: binary.LittleEndian.Uint32(data),
	}, nil
}
