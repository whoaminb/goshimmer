package mana

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestBalance_AddTransfer(t *testing.T) {
	calculator := NewCalculator(500, 0.1)

	// spend coins multiple times
	balance1 := NewBalance(calculator)
	balance1.AddTransfer(&Transfer{
		movedCoins:   1000,
		burnedMana:   10,
		receivedTime: 1000,
		spentTime:    1700,
	})
	balance1.AddTransfer(&Transfer{
		movedCoins:   1000,
		burnedMana:   0,
		receivedTime: 700,
		spentTime:    1000,
	})
	balance1.AddTransfer(&Transfer{
		movedCoins:   1000,
		burnedMana:   0,
		receivedTime: 0,
		spentTime:    700,
	})

	// hold coins for the full time
	balance2 := NewBalance(calculator)
	balance2.AddTransfer(&Transfer{
		movedCoins:   1000,
		burnedMana:   10,
		receivedTime: 0,
		spentTime:    1700,
	})

	// check result
	val1, _ := balance1.GetValue()
	assert.Equal(t, val1, uint64(291))
	val2, _ := balance2.GetValue()
	assert.Equal(t, val2, uint64(291))
}
