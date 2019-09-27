package social_consensus

import (
	"sync"
)

type TransactionDatabase struct {
	transactions map[int]*Transaction
	mutex        sync.RWMutex
}

func NewTransactionDatabase() *TransactionDatabase {
	return &TransactionDatabase{
		transactions: make(map[int]*Transaction),
	}
}

func (transactionDatabase *TransactionDatabase) StoreTransaction(transaction *Transaction) {
	transactionDatabase.mutex.Lock()
	defer transactionDatabase.mutex.Unlock()

	transactionDatabase.transactions[transaction.GetID()] = transaction
}

func (transactionDatabase *TransactionDatabase) LoadTransaction(id int) *Transaction {
	transactionDatabase.mutex.RLock()
	defer transactionDatabase.mutex.RUnlock()

	return transactionDatabase.transactions[id]
}
