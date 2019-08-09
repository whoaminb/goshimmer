package marshal

type Field interface {
	Get(obj interface{}) interface{}
	Marshal(obj interface{}, data []byte) []byte
}

type encoder struct {
	fields []Field
}

func Schema(fields ...Field) *encoder {
	return &encoder{
		fields: fields,
	}
}

func (encoder *encoder) Marshal(obj interface{}) []byte {
	result := make([]byte, 0)

	for _, field := range encoder.fields {
		result = append(result, field.Marshal(field.Get(obj), result)...)
	}

	return result
}
