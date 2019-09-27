package social_consensus

import (
	"math/rand"
	"time"

	"github.com/iotaledger/goshimmer/packages/events"
)

type Node struct {
	id                  int
	weight              int
	transactionCounter  int
	tipSelector         *TipSelector
	solidifier          *Solidifier
	approversDatabase   *ApproversDatabase
	transactionDatabase *TransactionDatabase
	gossipNeighbors     []*Node
	conflictSet         *ConflictSet
	Events              NodeEvents

	favoredReality int
}

var nodeIdCounter = 0

func NewNode(weight int) *Node {
	node := &Node{
		id:                  nodeIdCounter,
		weight:              weight,
		approversDatabase:   NewApproversDatabase(),
		transactionDatabase: NewTransactionDatabase(),
		gossipNeighbors:     make([]*Node, 0),
		conflictSet:         NewConflictSet(),
		Events: NodeEvents{
			TransactionSolid: events.NewEvent(TransactionEventHandler),
		},
	}

	node.Events.TransactionSolid.Attach(events.NewClosure(node.ClassifyTransaction))

	node.tipSelector = NewTipSelector(node)
	node.solidifier = NewSolidifier(node)

	nodeIdCounter++

	return node
}

func (node *Node) GetWeight() int {
	return node.weight
}

func (node *Node) GetID() int {
	return node.id
}

func (node *Node) Peer(nodes ...*Node) {
	node.gossipNeighbors = nodes
}

func (node *Node) GossipTransaction(transaction *Transaction) {
	node.ReceiveTransaction(transaction)

	for _, neighbor := range node.gossipNeighbors {
		go func(neighbor *Node, transaction *Transaction) {
			time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

			neighbor.ReceiveTransaction(transaction.Clone())
		}(neighbor, transaction)
	}
}

func (node *Node) StoreTransaction(transaction *Transaction) {
	if transaction != nil {
		node.transactionDatabase.StoreTransaction(transaction)
	}
}

func (node *Node) GetTransaction(id int) *Transaction {
	return node.transactionDatabase.LoadTransaction(id)
}

func (node *Node) CreateTransaction() *Transaction {
	trunk, branch, transactionLiked, realityLiked := node.tipSelector.GetTipsToApprove()

	newTransaction := NewTransaction()
	newTransaction.SetCounter(node.transactionCounter)
	newTransaction.SetNodeID(node.GetID())
	newTransaction.SetTrunkID(trunk.GetID())
	newTransaction.SetBranchID(branch.GetID())
	newTransaction.SetBranchTransactionLiked(transactionLiked)
	newTransaction.SetBranchRealityLiked(realityLiked)

	node.transactionCounter++

	return newTransaction
}

func (node *Node) IssueTransaction() {
	node.GossipTransaction(node.CreateTransaction())
}

func (node *Node) updateApprovers(transaction *Transaction) {
	if trunkID := transaction.GetTrunkID(); trunkID != -1 {
		approvers := node.approversDatabase.LoadApprovers(trunkID)
		if approvers == nil {
			approvers = NewApprovers(trunkID)

			node.approversDatabase.StoreApprovers(approvers)
		}

		approvers.AddTransactionLiker(transaction)
		approvers.AddRealityLiker(transaction)
	}

	if branchID := transaction.GetBranchID(); branchID != -1 {
		approvers := node.approversDatabase.LoadApprovers(branchID)
		if approvers == nil {
			approvers = NewApprovers(branchID)

			node.approversDatabase.StoreApprovers(approvers)
		}

		if transaction.IsBranchTransactionLiked() {
			approvers.AddTransactionLiker(transaction)
		} else {
			approvers.AddTransactionDisliker(transaction)
		}

		if transaction.IsBranchRealityLiked() {
			approvers.AddRealityLiker(transaction)
		} else {
			approvers.AddRealityDisliker(transaction)
		}
	}
}

func (node *Node) ReceiveTransaction(transaction *Transaction) {
	if node.transactionDatabase.LoadTransaction(transaction.GetID()) == nil {
		if transaction.claimedReality != 0 {
			node.conflictSet.AddReality(transaction.claimedReality)

			if node.favoredReality == 0 {
				node.favoredReality = transaction.claimedReality
			}
		}

		node.StoreTransaction(transaction)
		node.updateApprovers(transaction)

		node.solidifier.ProcessTransaction(transaction)
	}
}

func (node *Node) ClassifyTransaction(transaction *Transaction) {
	transactionLocallyLiked := transaction.claimedReality == node.favoredReality || transaction.claimedReality == 0
	realityLocallyLiked := transactionLocallyLiked && (transaction.GetMetadata().GetReality() == 0 || transaction.GetMetadata().GetReality() == node.favoredReality)

	transaction.GetMetadata().SetTransactionLocallyLiked(transactionLocallyLiked)
	transaction.GetMetadata().SetRealityLocallyLiked(realityLocallyLiked)

	node.tipSelector.ProcessClassifiedTransaction(transaction)
}
