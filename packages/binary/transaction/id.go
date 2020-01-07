package transaction

type Id [IdLength]byte

func (id *Id) MarshalBinary() (result []byte, err error) {
	result = make([]byte, IdLength)
	copy(result, id[:])

	return
}

func (id *Id) UnmarshalBinary(data []byte) (err error) {
	copy(id[:], data)

	return
}

var EmptyId = Id{}

const IdLength = 64
