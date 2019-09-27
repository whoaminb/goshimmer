package social_consensus

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/datastructure"
)

type TipSelector struct {
	node          *Node
	likedTips     *datastructure.RandomMap // like transaction / like reality
	semiLikedTips *datastructure.RandomMap // like transaction / dislike reality
	dislikedTips  *datastructure.RandomMap // dislike transaction / dislike reality
	// semiDislikedTips *datastructure.RandomMap // dislike transaction / like reality (doesn't occur: the only reason
	// to dislike a transaction is if it is conflicting. This however means it spawns a reality we can not both like and
	// dislike it at the same time.

	mutex sync.RWMutex
}

func NewTipSelector(node *Node) *TipSelector {
	return &TipSelector{
		node:          node,
		likedTips:     datastructure.NewRandomMap(),
		semiLikedTips: datastructure.NewRandomMap(),
		dislikedTips:  datastructure.NewRandomMap(),
	}
}

func (tipSelector *TipSelector) GetTipsToApprove() (trunkTransaction *Transaction, branchTransaction *Transaction, transactionLiked bool, realityLiked bool) {
	tipSelector.mutex.RLock()
	defer tipSelector.mutex.RUnlock()

	trunkTransaction = tipSelector.getRandomLikedTip()

	if branchTransaction = tipSelector.getRandomSemiLikedTip(); branchTransaction != nil {
		transactionLiked = true
		realityLiked = false

		return
	}

	if branchTransaction = tipSelector.getRandomDislikedTip(); branchTransaction != nil {
		transactionLiked = false
		realityLiked = false

		return
	}

	branchTransaction = tipSelector.getRandomLikedTip()
	transactionLiked = true
	realityLiked = true

	return
}

func (tipSelector *TipSelector) ProcessClassifiedTransaction(transaction *Transaction) {
	tipSelector.mutex.Lock()
	defer tipSelector.mutex.Unlock()

	transactionMetadata := transaction.GetMetadata()

	if transactionLiked := transactionMetadata.IsTransactionLocallyLiked(); transactionLiked {
		if realityLiked := transactionMetadata.IsRealityLocallyLiked(); realityLiked {
			tipSelector.likedTips.Set(transaction.GetID(), transaction)

			if trunk := tipSelector.node.GetTransaction(transaction.GetTrunkID()); trunk != nil {
				trunkId := trunk.GetID()

				tipSelector.likedTips.Delete(trunkId)
				tipSelector.semiLikedTips.Delete(trunkId)
				tipSelector.dislikedTips.Delete(trunkId)
			}

			if branch := tipSelector.node.GetTransaction(transaction.GetBranchID()); branch != nil {
				branchId := branch.GetID()

				tipSelector.likedTips.Delete(branchId)
				tipSelector.semiLikedTips.Delete(branchId)
				tipSelector.dislikedTips.Delete(branchId)
			}
		} else {
			tipSelector.semiLikedTips.Set(transaction.GetID(), transaction)
		}
	} else {
		if realityLiked := transactionMetadata.IsRealityLocallyLiked(); realityLiked {
			panic("we should never dislike a transaction and at the same time like its reality")
		} else {
			tipSelector.dislikedTips.Set(transaction.GetID(), transaction)
		}
	}

	//fmt.Println("TIP COUNT:", tipSelector.dislikedTips.Size()+tipSelector.semiLikedTips.Size()+tipSelector.likedTips.Size())
}

func (tipSelector *TipSelector) getRandomLikedTip() *Transaction {
	if result := tipSelector.likedTips.RandomEntry(); result == nil {
		return tipSelector.node.GetTransaction(0)
	} else {
		return result.(*Transaction)
	}
}

func (tipSelector *TipSelector) getRandomSemiLikedTip() *Transaction {
	if randomSemiLikedTip := tipSelector.semiLikedTips.RandomEntry(); randomSemiLikedTip != nil {
		return randomSemiLikedTip.(*Transaction)
	} else {
		return nil
	}
}

func (tipSelector *TipSelector) getRandomDislikedTip() *Transaction {
	if randomDislikedTip := tipSelector.dislikedTips.RandomEntry(); randomDislikedTip != nil {
		return randomDislikedTip.(*Transaction)
	} else {
		return nil
	}
}
