package ledgerstate

import (
	"testing"
)

func Benchmark(b *testing.B) {
	address1 := NewAddressHash("ADDRESS1")
	address2 := NewAddressHash("ADDRESS2")
	address3 := NewAddressHash("ADDRESS3")
	transfer1 := NewTransferHash("TRANSFER1")
	transfer2 := NewTransferHash("TRANSFER2")

	ledgerState := NewLedgerState([]byte("TESTLEDGER"))

	ledgerState.AddTransferOutputOld(
		NewTransferOutput(ledgerState, MAIN_REALITY_ID, address1, transfer1, NewColoredBalance("RED", 1337), NewColoredBalance("IOTA", 1338)),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		transfer := NewTransfer(transfer2).AddInput(
			NewTransferOutputReferenceOld(MAIN_REALITY_ID, address1, transfer1),
		).AddOutput(
			address2, NewColoredBalance("IOTA", 338),
		).AddOutput(
			address2, NewColoredBalance("RED", 337),
		).AddOutput(
			address3, NewColoredBalance("IOTA", 1000),
		).AddOutput(
			address3, NewColoredBalance("RED", 1000),
		)

		if err := ledgerState.BookTransfer(transfer); err != nil {
			panic(err)
		}
	}
}

func NewAddressHash(input string) (result AddressHash) {
	copy(result[:], input)
	return
}

func NewRealityId(input string) (result RealityId) {
	copy(result[:], input)
	return
}

func NewTransferHash(input string) (result TransferHash) {
	copy(result[:], input)
	return
}

func Test(t *testing.T) {
	//pendingReality := NewRealityId("PENDING")
	address1 := NewAddressHash("ADDRESS1")
	//address2 := NewAddressHash("ADDRESS2")
	//address3 := NewAddressHash("ADDRESS3")
	//address4 := NewAddressHash("ADDRESS4")
	transfer1 := NewTransferHash("TRANSFER1")
	//transfer2 := NewTransferHash("TRANSFER2")
	//transfer3 := NewTransferHash("TRANSFER3")

	ledgerState := NewLedgerState([]byte("TESTLEDGER"))

	ledgerState.AddTransferOutput(transfer1, address1, NewColoredBalance("RED", 1337), NewColoredBalance("IOTA", 1338))
	ledgerState.GetTransferOutput(NewTransferOutputReference(transfer1, address1))

	/*
		ledgerState.CreateReality(pendingReality)

		ledgerState.AddTransferOutputOld(
			NewTransferOutput(ledgerState, MAIN_REALITY_ID, address1, transfer1, NewColoredBalance("RED", 1337), NewColoredBalance("IOTA", 1338)),
		).AddTransferOutputOld(
			NewTransferOutput(ledgerState, pendingReality, address1, transfer1, NewColoredBalance("RED", 7331), NewColoredBalance("IOTA", 8331)),
		).AddTransferOutputOld(
			NewTransferOutput(ledgerState, pendingReality, address2, transfer2, NewColoredBalance("RED", 7331), NewColoredBalance("IOTA", 8331)),
		)

		fmt.Println(ledgerState.GetReality(pendingReality).GetAddress(address1).GetUnspentTransferOutputs())

		transfer := NewTransfer(transfer3).AddInput(
			NewTransferOutputReferenceOld(MAIN_REALITY_ID, address1, transfer1),
		).AddOutput(
			address3, NewColoredBalance("IOTA", 338),
		).AddOutput(
			address3, NewColoredBalance("RED", 337),
		).AddOutput(
			address4, NewColoredBalance("IOTA", 1000),
		).AddOutput(
			address4, NewColoredBalance("RED", 1000),
		)

		if err := ledgerState.BookTransfer(transfer); err != nil {
			t.Error(err)
		}

		fmt.Println(ledgerState.GetReality(MAIN_REALITY_ID).GetAddress(address3).GetBalances())
		fmt.Println(ledgerState.GetReality(MAIN_REALITY_ID).GetAddress(address4).GetBalances())
	*/
}
