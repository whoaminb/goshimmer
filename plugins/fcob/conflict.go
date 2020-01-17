package fcob

import (
	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/iota.go/trinary"
)

// getConflictSet triggers a (fake) new conflict if the tx's value is equal to 73798465
// including only 1 conflicting tx in the returned conflict set
func getConflictSet(transaction trinary.Trytes, tangle tangleAPI) (conflictSet map[trinary.Trytes]bool, err errors.IdentifiableError) {

	conflictSet = make(map[trinary.Trytes]bool)

	txObject, err := tangle.GetTransaction(transaction)
	if err != nil {
		return conflictSet, err
	}
	conflict := txObject.GetValue() == 73798465 // trigger a new conflict
	if conflict {
		conflictSet[transaction] = true
	}
	return conflictSet, nil
}
