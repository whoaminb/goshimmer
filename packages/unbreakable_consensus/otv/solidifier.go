package social_consensus

type Solidifier struct {
	node *Node
}

func NewSolidifier(node *Node) *Solidifier {
	return &Solidifier{
		node: node,
	}
}

func (solidifier *Solidifier) IsTransactionSolid(transaction *Transaction) (isSolid bool) {
	isSolid = true

	if trunkID := transaction.GetTrunkID(); trunkID != -1 {
		if trunk := solidifier.node.GetTransaction(trunkID); trunk == nil {
			isSolid = false
		} else if !trunk.GetMetadata().IsSolid() {
			isSolid = false
		}
	}

	if branchID := transaction.GetBranchID(); branchID != -1 {
		if trunk := solidifier.node.GetTransaction(branchID); trunk == nil {
			isSolid = false
		} else if !trunk.GetMetadata().IsSolid() {
			isSolid = false
		}
	}

	return
}

func (solidifier *Solidifier) PropagateSolidity(transaction *Transaction) {
	if !transaction.GetMetadata().IsSolid() {
		transaction.GetMetadata().SetSolid(true)

		if transaction.claimedReality != 0 {
			transaction.GetMetadata().SetReality(transaction.claimedReality)
		} else {
			trunk := solidifier.node.GetTransaction(transaction.GetTrunkID())
			trunkReality := trunk.GetMetadata().GetReality()
			branch := solidifier.node.GetTransaction(transaction.GetBranchID())
			branchReality := branch.GetMetadata().GetReality()

			if transaction.IsBranchRealityLiked() && branchReality != 0 {
				transaction.GetMetadata().SetReality(branchReality)

				solidifier.node.conflictSet.GetReality(branchReality).AddSupporter(transaction.GetNodeID(), transaction.counter)

				var heaviestReality *Reality
				heaviestRealityWeight := 0
				for _, reality := range solidifier.node.conflictSet.GetRealities() {
					if reality.weight > heaviestRealityWeight {
						heaviestReality = reality
						heaviestRealityWeight = reality.weight
					}
				}

				if solidifier.node.favoredReality != heaviestReality.id {
					solidifier.node.favoredReality = heaviestReality.id
				}
			} else {
				transaction.GetMetadata().SetReality(trunkReality)
			}
		}

		solidifier.node.Events.TransactionSolid.Trigger(transaction)

		if approvers := solidifier.node.approversDatabase.LoadApprovers(transaction.GetID()); approvers != nil {
			for _, approver := range approvers.transactionLikers {
				solidifier.PropagateSolidity(approver)
			}
			for _, approver := range approvers.transactionDislikers {
				solidifier.PropagateSolidity(approver)
			}
			for _, approver := range approvers.realityLikers {
				solidifier.PropagateSolidity(approver)
			}
			for _, approver := range approvers.realityDislikers {
				solidifier.PropagateSolidity(approver)
			}
		}
	}
}

func (solidifier *Solidifier) ProcessTransaction(transaction *Transaction) {
	if solidifier.IsTransactionSolid(transaction) {
		solidifier.PropagateSolidity(transaction)
	}
}
