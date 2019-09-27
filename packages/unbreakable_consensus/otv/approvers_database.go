package otv

import (
	"sync"
)

type ApproversDatabase struct {
	approvers map[int]*Approvers

	mutex sync.RWMutex
}

func NewApproversDatabase() *ApproversDatabase {
	return &ApproversDatabase{
		approvers: make(map[int]*Approvers),
	}
}

func (approversDatabase *ApproversDatabase) StoreApprovers(approvers *Approvers) {
	approversDatabase.mutex.Lock()
	defer approversDatabase.mutex.Unlock()

	approversDatabase.approvers[approvers.GetTransactionID()] = approvers
}

func (approversDatabase *ApproversDatabase) LoadApprovers(transactionID int) *Approvers {
	approversDatabase.mutex.RLock()
	defer approversDatabase.mutex.RUnlock()

	return approversDatabase.approvers[transactionID]
}
