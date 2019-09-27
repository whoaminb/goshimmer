package otv

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/iotaledger/goshimmer/packages/events"
)

func GetRealityFLC(node *Node, transaction *Transaction) (result []*Transaction) {
	result = make([]*Transaction, 1)
	result[0] = transaction

	traversalStack := make([]*Transaction, 0)
	seenTransactions := make(map[int]bool)

	if approvers := node.approversDatabase.LoadApprovers(transaction.GetID()); approvers != nil {
		for _, approver := range approvers.transactionLikers {
			if _, exists := seenTransactions[approver.GetID()]; !exists {
				traversalStack = append(traversalStack, approver)
				seenTransactions[approver.GetID()] = true
			}
		}

		for _, approver := range approvers.realityLikers {
			if _, exists := seenTransactions[approver.GetID()]; !exists {
				traversalStack = append(traversalStack, approver)
				seenTransactions[approver.GetID()] = true
			}
		}
	}

	for len(traversalStack) != 0 {
		currentTransaction := traversalStack[0]

		result = append(result, currentTransaction)

		if approvers := node.approversDatabase.LoadApprovers(currentTransaction.GetID()); approvers != nil {
			for _, approver := range approvers.realityLikers {
				if _, seen := seenTransactions[approver.GetID()]; !seen {
					traversalStack = append(traversalStack, approver)
					seenTransactions[approver.GetID()] = true
				}
			}
		}

		traversalStack = traversalStack[1:]
	}

	return result
}

func BenchmarkTPS(b *testing.B) {
	transactionIdCounter = 0

	genesisTransaction := NewTransaction()
	genesisTransaction.SetTrunkID(-1)
	genesisTransaction.SetBranchID(-1)
	genesisTransaction.GetMetadata().SetSolid(true)

	node := NewNode(100)
	node.StoreTransaction(genesisTransaction)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		node.IssueTransaction()
	}
}

func TestTipSelector_GetTipsToApprove(t *testing.T) {
	// region initialize nodes /////////////////////////////////////////////////////////////////////////////////////////

	genesisTransaction := NewTransaction()
	genesisTransaction.SetTrunkID(-1)
	genesisTransaction.SetBranchID(-1)
	genesisTransaction.GetMetadata().SetSolid(true)

	node0 := NewNode(15)
	node0.StoreTransaction(genesisTransaction)

	node1 := NewNode(17)
	node1.StoreTransaction(genesisTransaction)

	node2 := NewNode(19)
	node2.StoreTransaction(genesisTransaction)

	node0.Peer(node1, node2)
	node1.Peer(node0, node2)
	node2.Peer(node0, node1)

	// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////

	var wg sync.WaitGroup

	node0.Events.TransactionSolid.Attach(events.NewClosure(func(transaction *Transaction) {
		wg.Done()
	}))

	// region initialize consensus test conditions /////////////////////////////////////////////////////////////////////

	// create first conflicting transaction
	wg.Add(1)
	conflictingTransaction0 := node0.CreateTransaction()
	conflictingTransaction0.claimedReality = 1

	// create second conflicting transaction
	wg.Add(1)
	conflictingTransaction1 := node1.CreateTransaction()
	conflictingTransaction1.claimedReality = 2

	// make nodes see the conflicts in different order
	node0.GossipTransaction(conflictingTransaction0)
	node1.GossipTransaction(conflictingTransaction1)
	node2.GossipTransaction(conflictingTransaction0)

	// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////

	// issue a few transactions
	for i := 0; i < 30; i++ {
		wg.Add(3)

		node0.IssueTransaction()
		node1.IssueTransaction()
		node2.IssueTransaction()

		time.Sleep(50 * time.Millisecond)
	}

	wg.Wait()

	// check if consensus reached
	assert.Equal(t, node0.favoredReality, node1.favoredReality)
	assert.Equal(t, node1.favoredReality, node2.favoredReality)
}
