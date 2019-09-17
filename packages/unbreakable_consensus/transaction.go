package unbreakable_consensus

import (
	"github.com/iotaledger/goshimmer/packages/stringify"
)

type Transaction struct {
	id           string
	claimedEpoch int
}

func NewTransaction(id string) *Transaction {
	return &Transaction{
		id: id,
	}
}

func (transaction *Transaction) GetId() string {
	return transaction.id
}

func (transaction *Transaction) GetClaimedEpoch() int {
	return transaction.claimedEpoch
}

func (transaction *Transaction) String() string {
	return stringify.Struct("Transaction",
		stringify.StructField("id", transaction.id),
		stringify.StructField("claimedEpoch", transaction.claimedEpoch),
	)
}
