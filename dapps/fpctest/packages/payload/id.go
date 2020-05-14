package payload

import (
	"crypto/rand"
	"fmt"

	"github.com/iotaledger/hive.go/marshalutil"
	"github.com/mr-tron/base58"
)

// ID represents the hash of a payload that is used to identify the given payload.
type ID [IDLength]byte

// NewID creates a payload id from a base58 encoded string.
func NewID(base58EncodedString string) (result ID, err error) {
	bytes, err := base58.Decode(base58EncodedString)
	if err != nil {
		return
	}

	if len(bytes) != IDLength {
		err = fmt.Errorf("length of base58 formatted payload id is wrong")

		return
	}

	copy(result[:], bytes)

	return
}

// ParseID is a wrapper for simplified unmarshaling in a byte stream using the marshalUtil package.
func ParseID(marshalUtil *marshalutil.MarshalUtil) (ID, error) {
	id, err := marshalUtil.Parse(func(data []byte) (interface{}, int, error) { return IDFromBytes(data) })
	if err != nil {
		return ID{}, err
	}

	return id.(ID), nil
}

// IDFromBytes unmarshals a payload id from a sequence of bytes.
// It either creates a new payload id or fills the optionally provided object with the parsed information.
func IDFromBytes(bytes []byte, optionalTargetObject ...*ID) (result ID, consumedBytes int, err error) {
	// determine the target object that will hold the unmarshaled information
	var targetObject *ID
	switch len(optionalTargetObject) {
	case 0:
		targetObject = &result
	case 1:
		targetObject = optionalTargetObject[0]
	default:
		panic("too many arguments in call to IDFromBytes")
	}

	// initialize helper
	marshalUtil := marshalutil.New(bytes)

	// read id from bytes
	idBytes, err := marshalUtil.ReadBytes(IDLength)
	if err != nil {
		return
	}
	copy(targetObject[:], idBytes)

	// copy result if we have provided a target object
	result = *targetObject

	// return the number of bytes we processed
	consumedBytes = marshalUtil.ReadOffset()

	return
}

// RandomID creates a random payload id which can for example be used in unit tests.
func RandomID() (id ID) {
	// generate a random sequence of bytes
	idBytes := make([]byte, IDLength)
	if _, err := rand.Read(idBytes); err != nil {
		panic(err)
	}

	// copy the generated bytes into the result
	copy(id[:], idBytes)

	return
}

// String returns a base58 encoded version of the payload id.
func (id ID) String() string {
	return base58.Encode(id[:])
}

// Bytes returns a marshaled version of this ID.
func (id ID) Bytes() []byte {
	return id[:]
}

// GenesisID contains the zero value of this ID which represents the genesis.
var GenesisID ID

// IDLength defined the amount of bytes in a payload id (32 bytes hash value).
const IDLength = 32