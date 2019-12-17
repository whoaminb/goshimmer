package transaction

import (
	"encoding"
)

type Payload interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler

	GetType() int
}
