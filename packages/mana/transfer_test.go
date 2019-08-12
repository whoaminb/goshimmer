package mana

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestTransfer_Equals(t *testing.T) {
	// create test transfers
	transfer1 := NewTransfer([]*Input{}, 1337, 1338)
	transfer2 := NewTransfer([]*Input{}, 1337, 1339)
	transfer3 := NewTransfer([]*Input{}, 1337, 1338)
	transfer4 := NewTransfer([]*Input{}, 1339, 1338)
	transfer5 := NewTransfer([]*Input{NewInput(10, 10)}, 1337, 1338)
	transfer6 := NewTransfer([]*Input{NewInput(20, 10)}, 1337, 1338)
	transfer7 := NewTransfer([]*Input{NewInput(10, 10)}, 1337, 1338)

	// burned mana differs
	assert.Equal(t, transfer1.Equals(transfer2), false)

	// transfers are equal
	assert.Equal(t, transfer1.Equals(transfer3), true)

	// spentTime differs
	assert.Equal(t, transfer1.Equals(transfer4), false)

	// inputs length differs
	assert.Equal(t, transfer1.Equals(transfer5), false)

	// inputs differ
	assert.Equal(t, transfer5.Equals(transfer6), false)

	// transfers equal
	assert.Equal(t, transfer5.Equals(transfer7), true)
}

func TestTransfer_MarshalUnmarshalBinary(t *testing.T) {
	// create original transfer
	originalTransfer := NewTransfer([]*Input{}, 1337, 1338)

	// marshal
	marshaledTransfer, err := originalTransfer.MarshalBinary()
	if err != nil {
		t.Error(err)

		return
	}

	// unmarshal
	var unmarshaledTransfer Transfer
	if err := unmarshaledTransfer.UnmarshalBinary(marshaledTransfer); err != nil {
		t.Error(err)

		return
	}

	// compare result
	assert.Equal(t, unmarshaledTransfer.Equals(originalTransfer), true)
}
