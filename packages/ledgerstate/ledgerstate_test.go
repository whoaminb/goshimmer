package ledgerstate

import (
	"fmt"
	"testing"
)

func TestLedgerState_GetUnspentTransferOutputs(t *testing.T) {
	ledgerState := NewLedgerState()

	ledgerState.AddAddress(
		NewAddress("ABC").AddTransferOutput(
			NewTransferOutput("12").SetColoredBalance("I", 144).SetColoredBalance("A", 77),
		),
	)

	fmt.Println(ledgerState.GetUnspentTransferOutputs("ABC"))
}
