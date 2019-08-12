package mana

import (
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/iotaledger/goshimmer/packages/errors"
	manaproto "github.com/iotaledger/goshimmer/packages/mana/proto"
	"github.com/iotaledger/goshimmer/packages/marshaling"
)

type BalanceHistoryEntry struct {
	transfer      *Transfer
	transferMutex sync.RWMutex
	balance       uint64
	balanceMutex  sync.RWMutex
}

func (balanceHistoryEntry *BalanceHistoryEntry) GetTransfer() *Transfer {
	balanceHistoryEntry.transferMutex.RLock()
	defer balanceHistoryEntry.transferMutex.RUnlock()

	return balanceHistoryEntry.transfer
}

func (balanceHistoryEntry *BalanceHistoryEntry) SetTransfer(transfer *Transfer) {
	balanceHistoryEntry.transferMutex.Lock()
	defer balanceHistoryEntry.transferMutex.Unlock()

	balanceHistoryEntry.transfer = transfer
}

func (balanceHistoryEntry *BalanceHistoryEntry) GetBalance() uint64 {
	balanceHistoryEntry.balanceMutex.RLock()
	defer balanceHistoryEntry.balanceMutex.RUnlock()

	return balanceHistoryEntry.balance
}

func (balanceHistoryEntry *BalanceHistoryEntry) SetBalance(balance uint64) {
	balanceHistoryEntry.balanceMutex.Lock()
	defer balanceHistoryEntry.balanceMutex.Unlock()

	balanceHistoryEntry.balance = balance
}

func (balanceHistoryEntry *BalanceHistoryEntry) ToProto() proto.Message {
	balanceHistoryEntry.transferMutex.RLock()
	balanceHistoryEntry.balanceMutex.RLock()
	defer balanceHistoryEntry.transferMutex.RUnlock()
	defer balanceHistoryEntry.balanceMutex.RUnlock()

	return &manaproto.BalanceHistoryEntry{
		Transfer: balanceHistoryEntry.transfer.ToProto().(*manaproto.Transfer),
		Balance:  balanceHistoryEntry.balance,
	}
}

func (balanceHistoryEntry *BalanceHistoryEntry) FromProto(proto proto.Message) {
	balanceHistoryEntry.transferMutex.Lock()
	balanceHistoryEntry.balanceMutex.Lock()
	defer balanceHistoryEntry.transferMutex.Unlock()
	defer balanceHistoryEntry.balanceMutex.Unlock()

	protobufBalanceHistoryEntry := proto.(*manaproto.BalanceHistoryEntry)

	balanceHistoryEntry.balance = protobufBalanceHistoryEntry.Balance
	balanceHistoryEntry.transfer = &Transfer{}

	balanceHistoryEntry.transfer.FromProto(protobufBalanceHistoryEntry.Transfer)
}

func (balanceHistoryEntry *BalanceHistoryEntry) MarshalBinary() ([]byte, errors.IdentifiableError) {
	return marshaling.Marshal(balanceHistoryEntry)
}

func (balanceHistoryEntry *BalanceHistoryEntry) UnmarshalBinary(data []byte) (err errors.IdentifiableError) {
	return marshaling.Unmarshal(balanceHistoryEntry, data, &manaproto.BalanceHistoryEntry{})
}

func (balanceHistoryEntry *BalanceHistoryEntry) Equals(other *BalanceHistoryEntry) bool {
	if balanceHistoryEntry == other {
		return true
	}

	if balanceHistoryEntry == nil || other == nil {
		return false
	}

	balanceHistoryEntry.balanceMutex.RLock()
	balanceHistoryEntry.transferMutex.RLock()
	other.balanceMutex.RLock()
	other.transferMutex.RLock()
	defer balanceHistoryEntry.balanceMutex.RUnlock()
	defer balanceHistoryEntry.transferMutex.RUnlock()
	defer other.balanceMutex.RUnlock()
	defer other.transferMutex.RUnlock()

	return balanceHistoryEntry.balance == other.balance && balanceHistoryEntry.transfer.Equals(other.transfer)
}
