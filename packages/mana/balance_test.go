package mana

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestBalance_AddTransfer(t *testing.T) {
	calculator := NewCalculator(500, 0.1)

	// spend coins multiple times
	balance1 := NewBalance(calculator)
	balance1.AddTransfer(NewTransfer(1000, 1000, 1700))
	balance1.AddTransfer(NewTransfer(1000, 700, 1000))
	balance1.AddTransfer(NewTransfer(1000, 0, 700))

	// hold coins for the full time
	balance2 := NewBalance(calculator)
	balance2.AddTransfer(NewTransfer(1000, 0, 1700))

	// check result
	assert.Equal(t, balance1.GetValue(), uint64(301))
	assert.Equal(t, balance2.GetValue(), uint64(301))
}
