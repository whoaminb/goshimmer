package social_consensus

import (
	"sync"
)

type Reality struct {
	id          int
	conflictSet *ConflictSet
	supporters  map[int]int
	weight      int
	mutex       sync.RWMutex
}

func NewReality(conflictSet *ConflictSet, id int) *Reality {
	return &Reality{
		id:          id,
		conflictSet: conflictSet,
		supporters:  make(map[int]int),
	}
}

func (reality *Reality) GetId() int {
	return reality.id
}

func (reality *Reality) GetConflictSet() *ConflictSet {
	return reality.conflictSet
}

func (reality *Reality) ContainsSupporter(nodeId int) (transactionCounter int) {
	if currentTransactionCounter, supporterExists := reality.supporters[nodeId]; supporterExists {
		transactionCounter = currentTransactionCounter
	}

	return
}

func (reality *Reality) AddSupporter(nodeId int, transactionCounter int) {
	reality.mutex.Lock()
	defer reality.mutex.Unlock()

	if oldTransactionCounter, supporterExists := reality.supporters[nodeId]; !supporterExists {
		// remove supporter from the alternative reality
		for otherRealityId, otherReality := range reality.conflictSet.GetRealities() {
			if otherTransactionCounter := otherReality.ContainsSupporter(nodeId); otherRealityId != reality.id && otherTransactionCounter != 0 {
				otherReality.RemoveSupporter(nodeId, transactionCounter)
			}
		}

		// add supporter to new reality
		reality.supporters[nodeId] = transactionCounter

		// update weight
		reality.weight += 20
	} else if oldTransactionCounter < transactionCounter {
		// update transaction counter
		reality.supporters[nodeId] = transactionCounter
	}
}

func (reality *Reality) RemoveSupporter(nodeId int, transactionCounter int) {
	reality.mutex.Lock()
	defer reality.mutex.Unlock()

	if oldTransactionCounter, supporterExists := reality.supporters[nodeId]; supporterExists && oldTransactionCounter < transactionCounter {
		delete(reality.supporters, nodeId)

		reality.weight -= 20
	}
}
