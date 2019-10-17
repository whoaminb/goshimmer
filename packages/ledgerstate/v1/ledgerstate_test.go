package v1

import (
	"fmt"
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/transfer"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/coloredbalance"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/transferoutput"
)

func Benchmark(b *testing.B) {
	ledgerState := NewLedgerState([]byte("TESTLEDGER"))

	ledgerState.AddTransferOutput(
		transferoutput.New(ledgerState, MAIN_REALITY_ID, "ADDRESS1", "TRANSFER1", coloredbalance.New("RED", 1337), coloredbalance.New("IOTA", 1338)),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		transfer := transfer.New("TESTINGTON").AddInput(
			transferoutput.NewReference(MAIN_REALITY_ID, "ADDRESS1", "TRANSFER1"),
		).AddOutput(
			"ADDRESS4", coloredbalance.New("IOTA", 338),
		).AddOutput(
			"ADDRESS4", coloredbalance.New("RED", 337),
		).AddOutput(
			"ADDRESS5", coloredbalance.New("IOTA", 1000),
		).AddOutput(
			"ADDRESS5", coloredbalance.New("RED", 1000),
		)

		if err := ledgerState.BookTransfer(transfer); err != nil {
			panic(err)
		}
	}
}

func TestNewLedgerState(t *testing.T) {
	ledgerState := NewLedgerState([]byte("TESTLEDGER"))

	ledgerState.CreateReality("PENDING")

	ledgerState.AddTransferOutput(
		transferoutput.New(ledgerState, MAIN_REALITY_ID, "ADDRESS1", "TRANSFER1", coloredbalance.New("RED", 1337), coloredbalance.New("IOTA", 1338)),
	).AddTransferOutput(
		transferoutput.New(ledgerState, "PENDING", "ADDRESS1", "TRANSFER2", coloredbalance.New("RED", 7331), coloredbalance.New("IOTA", 8331)),
	).AddTransferOutput(
		transferoutput.New(ledgerState, "PENDING2", "ADDRESS2", "TRANSFER2", coloredbalance.New("RED", 7331), coloredbalance.New("IOTA", 8331)),
	)

	transfer := transfer.New("TESTINGTON").AddInput(
		transferoutput.NewReference(MAIN_REALITY_ID, "ADDRESS2", "TRANSFER1"),
	).AddOutput(
		"ADDRESS4", coloredbalance.New("IOTA", 338),
	).AddOutput(
		"ADDRESS4", coloredbalance.New("RED", 337),
	).AddOutput(
		"ADDRESS5", coloredbalance.New("IOTA", 1000),
	).AddOutput(
		"ADDRESS5", coloredbalance.New("RED", 1000),
	)

	if err := ledgerState.BookTransfer(transfer); err != nil {
		t.Error(err)
	}

	fmt.Println(ledgerState.GetReality(MAIN_REALITY_ID).GetAddress("ADDRESS4").GetBalances())
	fmt.Println(ledgerState.GetReality(MAIN_REALITY_ID).GetAddress("ADDRESS5").GetBalances())
}
