package ledgerstate

import (
	"fmt"
	"testing"
)

func TestLedgerState_GetUnspentTransferOutputs(t *testing.T) {
	ledgerState := NewLedgerState()

	ledgerState.GetReality().SetAddress(
		NewAddress("ABC").AddTransferOutput(
			"12", NewColoredBalance("I", 144),
		).AddTransferOutput(
			"12", NewColoredBalance("A", 77),
		).AddTransferOutput(
			"13", NewColoredBalance("I", 1000),
		),
	)

	/*
		ledgerState.AddAddress(
			NewAddress("ABC").AddTransferOutput(
				NewTransferOutput("12").SetColoredBalance("I", 144).SetColoredBalance("A", 77),
			).AddTransferOutput(
				NewTransferOutput("13").SetColoredBalance("I", 1000),
			),
		)
	*/

	transfer := NewTransfer("TEST").AddInputs("ABC", "12", "13").AddOutput(
		"CDE", NewColoredBalance("I", 144),
	).AddOutput(
		"GQL", NewColoredBalance("I", 1000),
	).AddOutput(
		"FGH", NewColoredBalance("A", 77),
	)

	fmt.Println(ledgerState.GetUnspentTransferOutputs("ABC"))

	fmt.Println(ledgerState.BookTransfer(transfer))

	fmt.Println(ledgerState.GetUnspentTransferOutputs("ABC"))
}
