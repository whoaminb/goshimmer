package transaction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/address"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/balance"
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
)

const ScIdLength = address.Length + balance.ColorLength

type ScId [ScIdLength]byte

func NewScId(addr *address.Address, color *balance.Color) *ScId {
	var ret ScId
	copy(ret[:address.Length], addr.Bytes())
	copy(ret[address.Length:], color.Bytes())
	return &ret
}

func (id *ScId) Bytes() []byte {
	return id[:]
}

func (id *ScId) Address() *address.Address {
	var ret address.Address
	copy(ret[:], id[:address.Length])
	return &ret
}

func (id *ScId) Color() *balance.Color {
	var ret balance.Color
	copy(ret[:], id[address.Length:])
	return &ret
}

func (id *ScId) String() string {
	return base58.Encode(id[:])
}

func (id *ScId) Short() string {
	return fmt.Sprintf("%s../%s..", id.Address().String()[:4], id.Color().String()[:4])
}

func ScIdFromString(s string) (*ScId, error) {
	b, err := base58.Decode(s)
	if err != nil {
		return nil, err
	}
	if len(b) != ScIdLength {
		return nil, errors.New("wrong hex encoded string. Can't convert to ScId")
	}
	var ret ScId
	copy(ret[:], b)
	return &ret, nil
}

func (id *ScId) Equal(id1 *ScId) bool {
	if id == id1 {
		return true
	}
	return bytes.Equal(id.Bytes(), id1.Bytes())
}

func (id *ScId) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

func (id *ScId) UnmarshalJSON(buf []byte) error {
	var s string
	err := json.Unmarshal(buf, &s)
	if err != nil {
		return err
	}
	ret, err := ScIdFromString(s)
	if err != nil {
		return err
	}
	copy(id.Bytes(), ret.Bytes())
	return nil
}
