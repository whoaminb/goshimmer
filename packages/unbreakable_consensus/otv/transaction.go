package otv

import (
	"github.com/iotaledger/goshimmer/packages/stringify"
)

type Transaction struct {
	id                     int
	nodeId                 int
	trunkId                int
	branchId               int
	counter                int
	branchRealityLiked     bool
	branchTransactionLiked bool

	metadata       *TransactionMetadata
	claimedReality int
}

var transactionIdCounter = 0

func NewTransaction() *Transaction {
	transaction := &Transaction{
		id: transactionIdCounter,
	}

	transactionIdCounter++

	return transaction
}

func (transaction *Transaction) SetNodeID(nodeID int) {
	transaction.nodeId = nodeID
}

func (transaction *Transaction) GetNodeID() int {
	return transaction.nodeId
}

func (transaction *Transaction) GetID() int {
	return transaction.id
}

func (transaction *Transaction) SetTrunkID(id int) {
	transaction.trunkId = id
}

func (transaction *Transaction) GetTrunkID() int {
	return transaction.trunkId
}

func (transaction *Transaction) SetBranchID(id int) {
	transaction.branchId = id
}

func (transaction *Transaction) GetBranchID() int {
	return transaction.branchId
}

func (transaction *Transaction) SetCounter(counter int) {
	transaction.counter = counter
}

func (transaction *Transaction) GetCounter() int {
	return transaction.counter
}

func (transaction *Transaction) SetBranchRealityLiked(liked bool) {
	transaction.branchRealityLiked = liked
}

func (transaction *Transaction) IsBranchRealityLiked() bool {
	return transaction.branchRealityLiked
}

func (transaction *Transaction) SetBranchTransactionLiked(liked bool) {
	transaction.branchTransactionLiked = liked
}

func (transaction *Transaction) IsBranchTransactionLiked() bool {
	return transaction.branchTransactionLiked
}

func (transaction *Transaction) GetMetadata() *TransactionMetadata {
	if transaction.metadata == nil {
		transaction.metadata = NewTransactionMetadata()
	}

	return transaction.metadata
}

func (transaction *Transaction) Clone() *Transaction {
	return &Transaction{
		id:                     transaction.id,
		nodeId:                 transaction.nodeId,
		trunkId:                transaction.GetTrunkID(),
		branchId:               transaction.GetBranchID(),
		counter:                transaction.counter,
		branchRealityLiked:     transaction.branchRealityLiked,
		branchTransactionLiked: transaction.branchTransactionLiked,
		metadata:               nil,
		claimedReality:         transaction.claimedReality,
	}
}

func (transaction *Transaction) String() string {
	return stringify.Struct("Transaction",
		stringify.StructField("id", transaction.GetID()),
		stringify.StructField("nodeID", transaction.GetNodeID()),
		stringify.StructField("trunkID", transaction.GetTrunkID()),
		stringify.StructField("branchID", transaction.GetBranchID()),
		stringify.StructField("branchTransactionLiked", transaction.IsBranchTransactionLiked()),
		stringify.StructField("branchRealityLiked", transaction.IsBranchRealityLiked()),
		stringify.StructField("claimedReality", transaction.claimedReality),
		stringify.StructField("reality", transaction.GetMetadata().GetReality()),
		stringify.StructField("isTransactionLocallyLiked", transaction.GetMetadata().IsTransactionLocallyLiked()),
		stringify.StructField("isRealityLocallyLiked", transaction.GetMetadata().IsRealityLocallyLiked()),
	)
}
