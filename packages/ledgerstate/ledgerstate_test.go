package ledgerstate

import (
	"fmt"
	"testing"
)

func Benchmark(b *testing.B) {
	ledgerState := NewLedgerState([]byte("TESTLEDGER"))

	ledgerState.AddTransferOutput(
		NewTransferOutput(ledgerState, MAIN_REALITY_ID, "ADDRESS1", "TRANSFER1", NewColoredBalance("RED", 1337), NewColoredBalance("IOTA", 1338)),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		transfer := NewTransfer("TESTINGTON").AddInput(
			NewTransferOutputReference(MAIN_REALITY_ID, "ADDRESS1", "TRANSFER1"),
		).AddOutput(
			"ADDRESS4", NewColoredBalance("IOTA", 338),
		).AddOutput(
			"ADDRESS4", NewColoredBalance("RED", 337),
		).AddOutput(
			"ADDRESS5", NewColoredBalance("IOTA", 1000),
		).AddOutput(
			"ADDRESS5", NewColoredBalance("RED", 1000),
		)

		if err := ledgerState.BookTransfer(transfer); err != nil {
			panic(err)
		}
	}
}

func Test(t *testing.T) {
	ledgerState := NewLedgerState([]byte("TESTLEDGER"))

	ledgerState.CreateReality("PENDING")

	ledgerState.AddTransferOutput(
		NewTransferOutput(ledgerState, MAIN_REALITY_ID, "ADDRESS1", "TRANSFER1", NewColoredBalance("RED", 1337), NewColoredBalance("IOTA", 1338)),
	).AddTransferOutput(
		NewTransferOutput(ledgerState, "PENDING", "ADDRESS1", "TRANSFER1", NewColoredBalance("RED", 7331), NewColoredBalance("IOTA", 8331)),
	).AddTransferOutput(
		NewTransferOutput(ledgerState, "PENDING", "ADDRESS2", "TRANSFER2", NewColoredBalance("RED", 7331), NewColoredBalance("IOTA", 8331)),
	)

	fmt.Println(ledgerState.GetReality("PENDING").GetAddress("ADDRESS1").GetUnspentTransferOutputs())

	transfer := NewTransfer("TESTINGTON").AddInput(
		NewTransferOutputReference(MAIN_REALITY_ID, "ADDRESS1", "TRANSFER1"),
	).AddOutput(
		"ADDRESS4", NewColoredBalance("IOTA", 338),
	).AddOutput(
		"ADDRESS4", NewColoredBalance("RED", 337),
	).AddOutput(
		"ADDRESS5", NewColoredBalance("IOTA", 1000),
	).AddOutput(
		"ADDRESS5", NewColoredBalance("RED", 1000),
	)

	if err := ledgerState.BookTransfer(transfer); err != nil {
		t.Error(err)
	}

	fmt.Println(ledgerState.GetReality(MAIN_REALITY_ID).GetAddress("ADDRESS4").GetBalances())
	fmt.Println(ledgerState.GetReality(MAIN_REALITY_ID).GetAddress("ADDRESS5").GetBalances())
}
