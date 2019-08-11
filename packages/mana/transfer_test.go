package mana

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestTransfer_MarshalUnmarshalBinary(t *testing.T) {
	// create original input
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
	assert.Equal(t, unmarshaledTransfer.spentTime, originalTransfer.spentTime)
	assert.Equal(t, unmarshaledTransfer.burnedMana, originalTransfer.burnedMana)
}
