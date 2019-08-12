package mana

import (
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/iotaledger/goshimmer/packages/datastructure"
	"github.com/iotaledger/goshimmer/packages/errors"
	manaproto "github.com/iotaledger/goshimmer/packages/mana/proto"
	"github.com/iotaledger/goshimmer/packages/marshaling"
)

type Balance struct {
	calculator     *Calculator
	balanceHistory *datastructure.DoublyLinkedList
	mutex          sync.RWMutex
}

func NewBalance(calculator *Calculator) *Balance {
	return &Balance{
		calculator:     calculator,
		balanceHistory: &datastructure.DoublyLinkedList{},
	}
}

// Returns the current mana balance.
func (balance *Balance) GetValue(now ...uint64) (result uint64, err errors.IdentifiableError) {
	balance.mutex.RLock()
	defer balance.mutex.RUnlock()

	if lastBalanceHistoryEntry, historyErr := balance.balanceHistory.GetLast(); historyErr != nil {
		if !datastructure.ErrNoSuchElement.Equals(historyErr) {
			err = historyErr
		}
	} else {
		switch len(now) {
		case 0:
			result = lastBalanceHistoryEntry.(*BalanceHistoryEntry).balance

		case 1:
			lastErosionTime := lastBalanceHistoryEntry.(*BalanceHistoryEntry).transfer.spentTime
			if lastErosionTime > now[0] {
				panic("watt")
			} else {
				result, _ = balance.calculator.ErodeMana(lastBalanceHistoryEntry.(*BalanceHistoryEntry).balance, now[0]-lastErosionTime)
			}

		default:
			err = errors.New("Test")
		}
	}

	return
}

// Returns the timestamp of the last mana erosion.
func (balance *Balance) GetLastErosion() uint64 {
	balance.mutex.RLock()
	defer balance.mutex.RUnlock()

	if lastBalanceHistoryEntry, err := balance.balanceHistory.GetLast(); datastructure.ErrNoSuchElement.Equals(err) {
		return 0
	} else {
		return lastBalanceHistoryEntry.(*BalanceHistoryEntry).transfer.spentTime
	}
}

// Books a new transfer to the balance.
func (balance *Balance) BookTransfer(transfer *Transfer) {
	balance.mutex.Lock()
	defer balance.mutex.Unlock()

	// check if we need to rollback transfers (to prevent rounding errors)
	rolledBackTransactions := balance.rollbackTransfers(transfer.spentTime)

	// apply new transfer
	balance.applyTransfer(transfer)

	// replay rolled back transfers (in reverse order)
	for i := len(rolledBackTransactions) - 1; i >= 0; i-- {
		balance.applyTransfer(rolledBackTransactions[i])
	}
}

// Cleans up old transfer history entries to reduce the size of the data.
func (balance *Balance) CleanupTransferHistory(referenceTime uint64) (err errors.IdentifiableError) {
	balance.mutex.Lock()
	defer balance.mutex.Unlock()

	if currentTransferHistoryEntry, historyErr := balance.balanceHistory.GetFirstEntry(); historyErr != nil {
		if !datastructure.ErrNoSuchElement.Equals(historyErr) {
			err = historyErr
		}
	} else {
		for currentTransferHistoryEntry.GetNext() != nil && currentTransferHistoryEntry.GetValue().(*BalanceHistoryEntry).transfer.spentTime < referenceTime {
			nextTransferHistoryEntry := currentTransferHistoryEntry.GetNext()

			if historyErr := balance.balanceHistory.RemoveEntry(currentTransferHistoryEntry); historyErr != nil {
				err = historyErr

				break
			}

			currentTransferHistoryEntry = nextTransferHistoryEntry
		}
	}

	return
}

func (balance *Balance) ToProto() proto.Message {
	balance.mutex.RLock()
	defer balance.mutex.RUnlock()

	balance.mutex.RLock()
	defer balance.mutex.RUnlock()

	balanceHistorySize := balance.balanceHistory.GetSize()
	protoBalance := &manaproto.Balance{
		BalanceHistory: make([]*manaproto.BalanceHistoryEntry, balanceHistorySize),
	}

	if balanceHistorySize >= 1 {
		if currentHistoryEntry, err := balance.balanceHistory.GetFirstEntry(); err != nil {
			panic(err)
		} else {
			for i := 0; i < balanceHistorySize; i++ {
				protoBalance.BalanceHistory[i] = currentHistoryEntry.GetValue().(*BalanceHistoryEntry).ToProto().(*manaproto.BalanceHistoryEntry)

				currentHistoryEntry = currentHistoryEntry.GetNext()
			}
		}
	}

	return protoBalance
}

func (balance *Balance) FromProto(proto proto.Message) {
	balance.mutex.Lock()
	defer balance.mutex.Unlock()

	balance.balanceHistory = &datastructure.DoublyLinkedList{}
	for _, balanceHistoryEntryProto := range proto.(*manaproto.Balance).BalanceHistory {
		var balanceHistoryEntry BalanceHistoryEntry
		balanceHistoryEntry.FromProto(balanceHistoryEntryProto)

		balance.balanceHistory.AddLast(&balanceHistoryEntry)
	}
}

func (balance *Balance) MarshalBinary() ([]byte, errors.IdentifiableError) {
	return marshaling.Marshal(balance)
}

func (balance *Balance) UnmarshalBinary(data []byte) (err errors.IdentifiableError) {
	return marshaling.Unmarshal(balance, data, &manaproto.Balance{})
}

func (balance *Balance) Equals(other *Balance) bool {
	if balance == other {
		return true
	}

	if balance == nil || other == nil {
		return false
	}

	balance.mutex.RLock()
	other.mutex.RLock()
	defer balance.mutex.RUnlock()
	defer other.mutex.RUnlock()

	if balance.balanceHistory == other.balanceHistory {
		return true
	}

	if balance.balanceHistory == nil || other.balanceHistory == nil {
		return false
	}

	if balance.balanceHistory.GetSize() != other.balanceHistory.GetSize() {
		return false
	}

	if balance.balanceHistory.GetSize() != 0 {
		ownTransferHistoryEntry, err := balance.balanceHistory.GetFirstEntry()
		if err != nil {
			// should never happen as we check the size before
			panic(err)
		}
		otherTransferHistoryEntry, err := other.balanceHistory.GetFirstEntry()
		if err != nil {
			// should never happen as we check the size before
			panic(err)
		}

		for ownTransferHistoryEntry != nil {
			if !ownTransferHistoryEntry.GetValue().(*BalanceHistoryEntry).Equals(otherTransferHistoryEntry.GetValue().(*BalanceHistoryEntry)) {
				return false
			}

			ownTransferHistoryEntry = ownTransferHistoryEntry.GetNext()
			otherTransferHistoryEntry = otherTransferHistoryEntry.GetNext()
		}
	}

	return true
}

// Rolls back transfers that have their spentTime after the given referenceTime and returns a slice containing the
// rolled back transfers.
//
// Since the mana calculations use floats, we will see rounding errors. To allow all nodes to have consensus on the
// current mana balance, we need to make nodes use the exact same formulas and apply them in the exact same order.
// Because of the asynchronous nature of the tangle, nodes will see different transactions at different times and will
// therefore process their mana gains in a different order. This could lead to discrepancies in the balance due to
// accumulated rounding errors. To work around this problem, we keep a history of the latest transfers (up till a
// certain age), that can be rolled back. This allows us to apply all mana changes in the exact same order which will
// lead to a network wide consensus on the mana balances.
func (balance *Balance) rollbackTransfers(referenceTime uint64) (result []*Transfer) {
	result = make([]*Transfer, 0)

	for {
		if lastListEntry, err := balance.balanceHistory.GetLast(); err != nil {
			if !datastructure.ErrNoSuchElement.Equals(err) {
				panic(err)
			}

			return
		} else if lastTransfer := lastListEntry.(*BalanceHistoryEntry).transfer; lastTransfer.spentTime < referenceTime {
			return
		} else {
			result = append(result, lastTransfer)

			if _, err := balance.balanceHistory.RemoveLast(); err != nil {
				panic(err)
			}
		}
	}
}

// Applies the balance changes of the given transfer.
func (balance *Balance) applyTransfer(transfer *Transfer) {
	// retrieve current values
	var currentBalance, lastErosion uint64
	if lastListEntry, err := balance.balanceHistory.GetLastEntry(); err != nil {
		if !datastructure.ErrNoSuchElement.Equals(err) {
			panic(err)
		}

		currentBalance = 0
		lastErosion = 0
	} else {
		lastBalanceHistoryEntry := lastListEntry.GetValue().(*BalanceHistoryEntry)

		currentBalance = lastBalanceHistoryEntry.balance
		lastErosion = lastBalanceHistoryEntry.transfer.spentTime
	}

	// erode if we have a balance
	if currentBalance != 0 {
		currentBalance, _ = balance.calculator.ErodeMana(currentBalance, transfer.spentTime-lastErosion)
	}

	// calculate mana gains
	var gainedMana uint64
	for _, input := range transfer.inputs {
		generatedMana, _ := balance.calculator.GenerateMana(input.GetCoinAmount(), transfer.spentTime-input.GetReceivedTime())

		gainedMana += generatedMana
	}

	// store results
	balance.balanceHistory.AddLast(&BalanceHistoryEntry{
		transfer: transfer,
		balance:  currentBalance + gainedMana - transfer.burnedMana,
	})
}
