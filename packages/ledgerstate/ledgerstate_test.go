package ledgerstate

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/iotaledger/goshimmer/packages/objectstorage"
)

var (
	iota           = NewColor("IOTA")
	eth            = NewColor("ETH")
	transferHash1  = NewTransferHash("TRANSFER1")
	transferHash2  = NewTransferHash("TRANSFER2")
	transferHash3  = NewTransferHash("TRANSFER3")
	addressHash1   = NewAddressHash("ADDRESS1")
	addressHash3   = NewAddressHash("ADDRESS3")
	addressHash4   = NewAddressHash("ADDRESS4")
	pendingReality = NewRealityId("PENDING")
)

func Benchmark(b *testing.B) {
	ledgerState := NewLedgerState("testLedger").Prune().AddTransferOutput(
		transferHash1, addressHash1, NewColoredBalance(eth, 1337),
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := ledgerState.BookTransfer(NewTransfer(NewTransferHash(strconv.Itoa(i))).AddInput(
			NewTransferOutputReference(transferHash1, addressHash1),
		).AddOutput(
			addressHash3, NewColoredBalance(eth, 1337),
		)); err != nil {
			b.Error(err)
		}
	}
}

func Test(t *testing.T) {
	ledgerState := NewLedgerState("testLedger").Prune().AddTransferOutput(
		transferHash1, addressHash1, NewColoredBalance(eth, 1337), NewColoredBalance(iota, 1338),
	)

	ledgerState.CreateReality(pendingReality)

	transfer := NewTransfer(transferHash2).AddInput(
		NewTransferOutputReference(transferHash1, addressHash1),
	).AddOutput(
		addressHash3, NewColoredBalance(iota, 338),
	).AddOutput(
		addressHash3, NewColoredBalance(eth, 337),
	).AddOutput(
		addressHash4, NewColoredBalance(iota, 1000),
	).AddOutput(
		addressHash4, NewColoredBalance(eth, 1000),
	)

	if err := ledgerState.BookTransfer(transfer); err != nil {
		t.Error(err)
	}

	if err := ledgerState.BookTransfer(NewTransfer(transferHash3).AddInput(
		NewTransferOutputReference(transferHash1, addressHash1),
	).AddOutput(
		addressHash3, NewColoredBalance(iota, 338),
	).AddOutput(
		addressHash3, NewColoredBalance(eth, 337),
	).AddOutput(
		addressHash4, NewColoredBalance(iota, 1000),
	).AddOutput(
		addressHash4, NewColoredBalance(eth, 1000),
	)); err != nil {
		t.Error(err)
	}

	objectstorage.WaitForWritesToFlush()

	ledgerState.ForEachTransferOutput(func(object *objectstorage.CachedObject) bool {
		object.Consume(func(object objectstorage.StorableObject) {
			fmt.Println(object.(*TransferOutput))
		})

		return true
	})
}
