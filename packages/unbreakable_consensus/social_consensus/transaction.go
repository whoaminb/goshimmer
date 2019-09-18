package social_consensus

import (
	"math/bits"

	"github.com/iotaledger/goshimmer/packages/stringify"
)

var transactionCounter = 0

type Transaction struct {
	id                      int
	claimedRound            int
	referencedElders        ElderMask
	seenSuperMajorityElders ElderMask
	issuer                  *Node
	elder                   *Node
	branch                  *Transaction
	trunk                   *Transaction
}

func NewTransaction() (result *Transaction) {
	result = &Transaction{
		id: transactionCounter,
	}

	transactionCounter++

	return
}

func (transaction *Transaction) SetIssuer(issuer *Node) {
	transaction.issuer = issuer
}

func (transaction *Transaction) SetElder(elder *Node) {
	transaction.elder = elder
}

func (transaction *Transaction) Attach(branch *Transaction, trunk *Transaction) {
	// update referenced elders
	transaction.referencedElders = branch.referencedElders.Union(trunk.referencedElders)
	if transaction.issuer != nil {
		transaction.referencedElders = transaction.referencedElders.Union(transaction.issuer.GetElderMask())
	}
	if transaction.elder != nil {
		transaction.referencedElders = transaction.referencedElders.Union(transaction.elder.GetElderMask())
	}

	// update claimed round
	if branch.claimedRound >= trunk.claimedRound {
		transaction.claimedRound = branch.claimedRound
	} else {
		transaction.claimedRound = trunk.claimedRound
	}

	if bits.OnesCount64(uint64(transaction.referencedElders)) > 14 {
		transaction.claimedRound++
		transaction.referencedElders = branch.referencedElders.Union(trunk.referencedElders)
	}

	// link transactions together
	transaction.branch = branch
	transaction.trunk = trunk
}

func (transaction *Transaction) String() string {
	return stringify.Struct("Transaction",
		stringify.StructField("id", transaction.id),
		stringify.StructField("issuer", transaction.issuer),
		stringify.StructField("elder", transaction.elder),
		stringify.StructField("claimedRound", transaction.claimedRound),
		stringify.StructField("referencedElders", transaction.referencedElders),
	)
}
