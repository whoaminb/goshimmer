package transaction

import (
	"bytes"
	"crypto/rand"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/address"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/balance"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	addr := address.RandomOfType(address.VERSION_BLS)
	blk := make([]byte, balance.ColorLength)
	_, err := rand.Read(blk)
	assert.Equal(t, err, nil)

	var color balance.Color
	copy(color[:], blk)
	scid := NewScId(&addr, &color)

	scidstr := scid.String()
	scid1, err := ScIdFromString(scidstr)
	assert.Equal(t, err, nil)
	assert.Equal(t, scidstr, scid1.String())

	assert.Equal(t, bytes.Equal(scid.Address().Bytes(), addr[:]), true)
	assert.Equal(t, bytes.Equal(scid.Color().Bytes(), color[:]), true)
}
