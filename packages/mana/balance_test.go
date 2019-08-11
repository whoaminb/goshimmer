package mana

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestBalance_CleanupTransferHistory(t *testing.T) {
	// initialize calculator
	calculator := NewCalculator(500, 0.1)

	// fill transfer history
	balance1 := NewBalance(calculator)
	balance1.BookTransfer(&Transfer{
		inputs: []*Input{NewInput(1000, 1000)},
		spentTime:  1700,
		burnedMana: 10,
	})
	balance1.BookTransfer(&Transfer{
		inputs: []*Input{NewInput(1000, 700)},
		spentTime:  1000,
		burnedMana: 0,
	})
	balance1.BookTransfer(&Transfer{
		inputs: []*Input{NewInput(1000, 0)},
		spentTime:  700,
		burnedMana: 0,
	})

	// cleanup transfer history
	if err := balance1.CleanupTransferHistory(1900); err != nil {
		t.Error(err)

		return
	}

	// check result (correct balance, correct history size)
	if val1, err := balance1.GetValue(); err != nil {
		t.Error(err)

		return
	} else {
		assert.Equal(t, val1, uint64(290))
	}
	assert.Equal(t, balance1.transferHistory.GetSize(), 1)
}

func TestBalance_AddTransfer(t *testing.T) {
	// initialize calculator
	calculator := NewCalculator(500, 0.1)

	// spend coins multiple times
	balance1 := NewBalance(calculator)
	balance1.BookTransfer(&Transfer{
		inputs: []*Input{NewInput(1000, 1000)},
		spentTime:  1700,
		burnedMana: 10,
	})
	balance1.BookTransfer(&Transfer{
		inputs: []*Input{NewInput(1000, 700)},
		spentTime:  1000,
		burnedMana: 0,
	})
	balance1.BookTransfer(&Transfer{
		inputs: []*Input{NewInput(1000, 0)},
		spentTime:  700,
		burnedMana: 0,
	})

	// hold coins for the full time
	balance2 := NewBalance(calculator)
	balance2.BookTransfer(&Transfer{
		inputs: []*Input{NewInput(1000, 0)},
		spentTime:  1700,
		burnedMana: 10,
	})

	// check result
	if val1, err := balance1.GetValue(); err != nil {
		t.Error(err)

		return
	} else {
		assert.Equal(t, val1, uint64(290))
	}
	if val2, err := balance2.GetValue(); err != nil {
		t.Error(err)

		return
	} else {
		assert.Equal(t, val2, uint64(291))
	}
}
