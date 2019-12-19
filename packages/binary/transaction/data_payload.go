package transaction

type DataPayload struct {
	payloadType PayloadType
	data        []byte
}

var DataPayloadType = PayloadType(0)

func NewDataPayload(data []byte) *DataPayload {
	return &DataPayload{
		payloadType: DataPayloadType,
		data:        data,
	}
}

func (dataPayload *DataPayload) GetType() PayloadType {
	return dataPayload.payloadType
}

func (dataPayload *DataPayload) GetData() []byte {
	return dataPayload.data
}

func (dataPayload *DataPayload) UnmarshalBinary(data []byte) error {
	dataPayload.data = make([]byte, len(data))
	copy(dataPayload.data, data)

	return nil
}

func (dataPayload *DataPayload) MarshalBinary() (data []byte, err error) {
	data = make([]byte, len(dataPayload.data))
	copy(data, dataPayload.data)

	return
}

func createGenericDataPayloadUnmarshaler(payloadType PayloadType) PayloadUnmarshaler {
	return func(data []byte) (payload Payload, err error) {
		payload = &DataPayload{
			payloadType: payloadType,
		}
		err = payload.UnmarshalBinary(data)

		return
	}
}
