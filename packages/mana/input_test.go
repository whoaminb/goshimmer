package mana

import (
	"testing"

	"github.com/magiconair/properties/assert"
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
	assert.Equal(t, unmarshaledInput.Equals(originalInput), true)
}

func TestInput_Equals(t *testing.T) {
	// create test transfers
	var input0 *Input
	input1 := NewInput(10, 12)
	input2 := NewInput(10, 14)
	input3 := NewInput(11, 12)
	input4 := NewInput(10, 12)

	// check results of Equals
	assert.Equal(t, input0.Equals(nil), true)
	assert.Equal(t, input1.Equals(nil), false)
	assert.Equal(t, input1.Equals(input2), false)
	assert.Equal(t, input1.Equals(input3), false)
	assert.Equal(t, input1.Equals(input4), true)
}
