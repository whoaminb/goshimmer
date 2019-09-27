package social_consensus

import (
	"sync"
)

type Approvers struct {
	transactionID        int
	transactionLikers    map[int]*Transaction
	transactionDislikers map[int]*Transaction
	realityLikers        map[int]*Transaction
	realityDislikers     map[int]*Transaction

	mutex sync.RWMutex
}

func NewApprovers(transactionID int) *Approvers {
	return &Approvers{
		transactionID:        transactionID,
		transactionLikers:    make(map[int]*Transaction),
		transactionDislikers: make(map[int]*Transaction),
		realityLikers:        make(map[int]*Transaction),
		realityDislikers:     make(map[int]*Transaction),
	}
}

func (approvers *Approvers) GetTransactionID() int {
	return approvers.transactionID
}

func (approvers *Approvers) AddTransactionLiker(transaction *Transaction) {
	approvers.mutex.Lock()
	defer approvers.mutex.Unlock()

	approvers.transactionLikers[transaction.GetID()] = transaction
}

func (approvers *Approvers) AddTransactionDisliker(transaction *Transaction) {
	approvers.mutex.Lock()
	defer approvers.mutex.Unlock()

	approvers.transactionDislikers[transaction.GetID()] = transaction
}

func (approvers *Approvers) AddRealityLiker(transaction *Transaction) {
	approvers.mutex.Lock()
	defer approvers.mutex.Unlock()

	approvers.realityLikers[transaction.GetID()] = transaction
}

func (approvers *Approvers) AddRealityDisliker(transaction *Transaction) {
	approvers.mutex.Lock()
	defer approvers.mutex.Unlock()

	approvers.realityDislikers[transaction.GetID()] = transaction
}
