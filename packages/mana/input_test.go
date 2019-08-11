package mana

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestInput_MarshalUnmarshalBinary(t *testing.T) {
	// create original input
	originalInput := NewInput(1337, 1338)

	// marshal
	marshaledInput, err := originalInput.MarshalBinary()
	if err != nil {
		t.Error(err)

		return
	}

	// unmarshal
	var unmarshaledInput Input
	if err := unmarshaledInput.UnmarshalBinary(marshaledInput); err != nil {
		t.Error(err)

		return
	}

	// compare result
	assert.Equal(t, unmarshaledInput.GetCoinAmount(), originalInput.GetCoinAmount())
	assert.Equal(t, unmarshaledInput.GetReceivedTime(), originalInput.GetReceivedTime())
}
