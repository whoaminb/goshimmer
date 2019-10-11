package ledgerstate

import (
	"fmt"
	"testing"
)

func TestLedgerState_GetUnspentTransferOutputs(t *testing.T) {
	ledgerState := NewLedgerState()

	NewAddress("ABC")

	fmt.Println(ledgerState.GetUnspentTransferOutputs("ABC"))
}
