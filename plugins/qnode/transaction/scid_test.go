package transaction

import (
	"bytes"
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/address"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	addr := address.RandomOfType(address.VERSION_BLS)
	color := RandomColor()
	scid := NewScId(&addr, color)

	scidstr := scid.String()
	scid1, err := ScIdFromString(scidstr)
	assert.Equal(t, err, nil)
	assert.Equal(t, scidstr, scid1.String())

	assert.Equal(t, bytes.Equal(scid.Address().Bytes(), addr[:]), true)
	assert.Equal(t, bytes.Equal(scid.Color().Bytes(), color[:]), true)
}

const (
	testAddress = "kKELws7qgMmpsufwf13CEQkRmYbCnrTg7f1qKNRgyVZ7"
	testScid    = "DsHiYnydheNLfhkc9sYPySVcEnyhxgtP4wWhKsczbnrRpYXrabwEjuej2N7bvb1qtdgGMewiWonzsD1zmLJAAXdE"
)

func TestRandScid(t *testing.T) {
	addr, err := address.FromBase58(testAddress)
	assert.Equal(t, err, nil)
	assert.Equal(t, addr.Version(), address.VERSION_BLS)

	scid := RandomScId(&addr)
	assert.Equal(t, bytes.Equal(scid.Address().Bytes(), addr[:]), true)
	t.Logf("scid = %s", scid.String())

	scid1, err := ScIdFromString(testScid)
	assert.Equal(t, err, nil)
	assert.Equal(t, scid.Equal(scid1), true)

}
