package transaction

import (
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
	copy(ret.Address().Bytes(), addr.Bytes())
	copy(ret.Color().Bytes(), color.Bytes())
	return &ret
}

func (id *ScId) Bytes() []byte {
	return id[:]
}

func (id *ScId) Address() (ret address.Address) {
	copy(ret.Bytes(), id.Bytes()[:address.Length])
	return
}

func (id *ScId) Color() (ret balance.Color) {
	copy(ret.Bytes(), id.Bytes()[address.Length:])
	return
}

func (id *ScId) String() string {
	return base58.Encode(id[:])
}

func (id *ScId) Short() string {
	return fmt.Sprintf("%s../%s..", id.Address().String()[:4])
}

func ScIdFromString(s string) (*ScId, error) {
	b, err := base58.Decode(s)
	//b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if len(b) != ScIdLength {
		return nil, errors.New("wrong hex encoded string. Can't convert to ScId")
	}
	var ret ScId
	copy(ret.Bytes(), b)
	return &ret, nil
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
