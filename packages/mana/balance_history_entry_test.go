package mana

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestBalanceHistoryEntry_MarshalUnmarshalBinary(t *testing.T) {
	balanceHistoryEntry := &BalanceHistoryEntry{
		transfer: NewTransfer([]*Input{NewInput(10, 100)}, 150, 10),
		balance:  100,
	}

	marshaledBalanceHistoryEntry, err := balanceHistoryEntry.MarshalBinary()
	if err != nil {
		t.Error(err)

		return
	}

	var unmarshaledBalanceHistoryEntry BalanceHistoryEntry
	unmarshaledBalanceHistoryEntry.UnmarshalBinary(marshaledBalanceHistoryEntry)

	assert.Equal(t, unmarshaledBalanceHistoryEntry.Equals(balanceHistoryEntry), true)
}
