package address

import (
	"crypto/rand"

	"github.com/mr-tron/base58"
)

type Address [Length]byte

func New(addressBytes []byte) (result Address) {
	copy(result[:], addressBytes)

	return
}

func FromBase58EncodedString(base58EncodedString string) (result Address) {
	if addressBytes, err := base58.Decode(base58EncodedString); err != nil {
		panic(err)
	} else {
		copy(result[:], addressBytes)
	}

	return
}

func Random() (result Address) {
	addressBytes := make([]byte, Length)
	if _, err := rand.Read(addressBytes); err != nil {
		panic(err)
	}

	return New(addressBytes)
}

func (address *Address) UnmarshalBinary(data []byte) error {
	copy(address[:], data[:Length])

	return nil
}

func (address *Address) MarshalBinary() (bytes []byte, err error) {
	bytes = append(bytes, address[:]...)

	return
}

func (address Address) String() string {
	return base58.Encode(address[:])
}

const Length = 32
