package ledgerstate

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/coloredcoins"

	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
	"github.com/iotaledger/hive.go/objectstorage"
	"github.com/iotaledger/hive.go/parameter"
)

var (
	iota_          = coloredcoins.NewColor("IOTA")
	eth            = coloredcoins.NewColor("ETH")
	transferHash1  = transfer.NewHash("TRANSFER1")
	transferHash2  = transfer.NewHash("TRANSFER2")
	transferHash3  = transfer.NewHash("TRANSFER3")
	transferHash4  = transfer.NewHash("TRANSFER4")
	transferHash5  = transfer.NewHash("TRANSFER5")
	transferHash6  = transfer.NewHash("TRANSFER6")
	addressHash1   = address.New([]byte("ADDRESS1"))
	addressHash2   = address.New([]byte("ADDRESS2"))
	addressHash3   = address.New([]byte("ADDRESS3"))
	addressHash4   = address.New([]byte("ADDRESS4"))
	addressHash5   = address.New([]byte("ADDRESS5"))
	addressHash6   = address.New([]byte("ADDRESS6"))
	pendingReality = reality.NewId("PENDING")
)

func init() {
	if err := parameter.FetchConfig(false); err != nil {
		panic(err)
	}
}

func Benchmark(b *testing.B) {
	ledgerState := NewLedgerState("testLedger").Prune().AddTransferOutput(
		transferHash1, addressHash1, coloredcoins.NewColoredBalance(eth, 1024),
	)

	b.ResetTimer()

	lastTransferHash := transferHash1

	for i := 0; i < b.N; i++ {
		newTransferHash := transfer.NewHash(strconv.Itoa(i))

		if err := ledgerState.BookTransfer(transfer.NewTransfer(newTransferHash).AddInput(
			transfer.NewOutputReference(lastTransferHash, addressHash1),
		).AddOutput(
			addressHash1, coloredcoins.NewColoredBalance(eth, 1024),
		)); err != nil {
			b.Error(err)
		}

		lastTransferHash = newTransferHash
	}
}

func Test(t *testing.T) {
	ledgerState := NewLedgerState("testLedger").Prune().AddTransferOutput(
		transferHash1, addressHash1, coloredcoins.NewColoredBalance(eth, 1337), coloredcoins.NewColoredBalance(iota_, 1338),
	)

	ledgerState.CreateReality(pendingReality)

	transfer1 := transfer.NewTransfer(transferHash2).AddInput(
		transfer.NewOutputReference(transferHash1, addressHash1),
	).AddOutput(
		addressHash3, coloredcoins.NewColoredBalance(iota_, 338),
	).AddOutput(
		addressHash3, coloredcoins.NewColoredBalance(eth, 337),
	).AddOutput(
		addressHash4, coloredcoins.NewColoredBalance(iota_, 1000),
	).AddOutput(
		addressHash4, coloredcoins.NewColoredBalance(eth, 1000),
	)

	if err := ledgerState.BookTransfer(transfer1); err != nil {
		t.Error(err)
	}

	if err := ledgerState.BookTransfer(transfer.NewTransfer(transferHash3).AddInput(
		transfer.NewOutputReference(transferHash1, addressHash1),
	).AddOutput(
		addressHash3, coloredcoins.NewColoredBalance(iota_, 338),
	).AddOutput(
		addressHash3, coloredcoins.NewColoredBalance(eth, 337),
	).AddOutput(
		addressHash4, coloredcoins.NewColoredBalance(iota_, 1000),
	).AddOutput(
		addressHash4, coloredcoins.NewColoredBalance(eth, 1000),
	)); err != nil {
		t.Error(err)
	}

	time.Sleep(1000 * time.Millisecond)

	objectstorage.WaitForWritesToFlush()

	ledgerState.ForEachTransferOutput(func(object *objectstorage.CachedObject) bool {
		object.Consume(func(object objectstorage.StorableObject) {
			fmt.Println(object.(*transfer.Output))
		})

		return true
	})
}

var transferHashCounter = 0

func generateRandomTransferHash() transfer.Hash {
	transferHashCounter++

	return transfer.NewHash("TRANSFER" + strconv.Itoa(transferHashCounter))
}

var addressHashCounter = 0

func generateRandomAddressHash() address.Address {
	addressHashCounter++

	return address.New([]byte("ADDRESS" + strconv.Itoa(addressHashCounter)))
}

func initializeLedgerStateWithBalances(numberOfBalances int) (ledgerState *LedgerState, result []*transfer.OutputReference) {
	ledgerState = NewLedgerState("testLedger").Prune()

	for i := 0; i < numberOfBalances; i++ {
		transferHash := generateRandomTransferHash()
		addressHash := generateRandomAddressHash()

		ledgerState.AddTransferOutput(transferHash, addressHash, coloredcoins.NewColoredBalance(iota_, 1024))

		result = append(result, transfer.NewOutputReference(transferHash, addressHash))
	}

	return
}

func doubleSpend(ledgerState *LedgerState, transferOutputReference *transfer.OutputReference) (result []*transfer.OutputReference) {
	for i := 0; i < 2; i++ {
		result = append(result, spend(ledgerState, transferOutputReference))
	}

	return
}

func spend(ledgerState *LedgerState, transferOutputReferences ...*transfer.OutputReference) (result *transfer.OutputReference) {
	transferHash := generateRandomTransferHash()
	addressHash := generateRandomAddressHash()

	totalInputBalance := uint64(0)

	transfer1 := transfer.NewTransfer(transferHash)
	for _, transferOutputReference := range transferOutputReferences {
		ledgerState.GetTransferOutput(transferOutputReference).Consume(func(object objectstorage.StorableObject) {
			transferOutput := object.(*transfer.Output)

			for _, coloredBalance := range transferOutput.GetBalances() {
				totalInputBalance += coloredBalance.GetBalance()
			}
		})

		transfer1.AddInput(transferOutputReference)
	}
	transfer1.AddOutput(
		addressHash, coloredcoins.NewColoredBalance(iota_, totalInputBalance),
	)

	if err := ledgerState.BookTransfer(transfer1); err != nil {
		panic(err)
	}

	result = transfer.NewOutputReference(transferHash, addressHash)

	return
}

func multiSpend(ledgerState *LedgerState, outputCount int, transferOutputReferences ...*transfer.OutputReference) (result []*transfer.OutputReference) {
	transferHash := generateRandomTransferHash()

	transfer1 := transfer.NewTransfer(transferHash)

	totalInputBalance := uint64(0)
	for _, transferOutputReference := range transferOutputReferences {
		ledgerState.GetTransferOutput(transferOutputReference).Consume(func(object objectstorage.StorableObject) {
			transferOutput := object.(*transfer.Output)

			for _, coloredBalance := range transferOutput.GetBalances() {
				totalInputBalance += coloredBalance.GetBalance()
			}
		})

		transfer1.AddInput(transferOutputReference)
	}

	for i := 0; i < outputCount; i++ {
		addressHash := generateRandomAddressHash()

		transfer1.AddOutput(
			addressHash, coloredcoins.NewColoredBalance(iota_, totalInputBalance/uint64(outputCount)),
		)

		result = append(result, transfer.NewOutputReference(transferHash, addressHash))
	}

	if err := ledgerState.BookTransfer(transfer1); err != nil {
		panic(err)
	}

	return
}

func TestAggregateAggregatedRealities(t *testing.T) {
	ledgerState, transferOutputs := initializeLedgerStateWithBalances(3)

	outputs0 := multiSpend(ledgerState, 2, multiSpend(ledgerState, 1, transferOutputs[0])[0])
	multiSpend(ledgerState, 1, transferOutputs[0])

	outputs1 := multiSpend(ledgerState, 2, multiSpend(ledgerState, 1, transferOutputs[1])[0])
	multiSpend(ledgerState, 1, transferOutputs[1])

	outputs2 := multiSpend(ledgerState, 2, multiSpend(ledgerState, 1, transferOutputs[2])[0])
	multiSpend(ledgerState, 1, transferOutputs[2])

	aggregatedOutputs0 := multiSpend(ledgerState, 2, outputs0[0], outputs1[0])
	aggregatedOutputs1 := multiSpend(ledgerState, 2, outputs1[1], outputs2[1])
	aggregatedOutputs2 := multiSpend(ledgerState, 2, outputs0[1], outputs2[0])

	multiSpend(ledgerState, 1, aggregatedOutputs0[0], aggregatedOutputs1[0])
	multiSpend(ledgerState, 1, aggregatedOutputs0[1], aggregatedOutputs2[0])

	time.Sleep(2000 * time.Millisecond)

	objectstorage.WaitForWritesToFlush()

	_ = ledgerState.GenerateRealityVisualization("realities1.png")
	_ = NewVisualizer(ledgerState).RenderTransferOutputs("outputs1.png")

	multiSpend(ledgerState, 2, outputs0[0], outputs1[0])

	time.Sleep(2000 * time.Millisecond)

	objectstorage.WaitForWritesToFlush()

	_ = ledgerState.GenerateRealityVisualization("realities2.png")
	_ = NewVisualizer(ledgerState).RenderTransferOutputs("outputs2.png")
}

func TestElevateAggregatedReality(t *testing.T) {
	ledgerState, transferOutputs := initializeLedgerStateWithBalances(3)

	// create 2 double spends
	doubleSpentOutputs1 := doubleSpend(ledgerState, transferOutputs[0])
	doubleSpentOutputs2 := doubleSpend(ledgerState, transferOutputs[1])
	normalSpend := spend(ledgerState, transferOutputs[2])
	doubleSpentOutputs3 := doubleSpend(ledgerState, normalSpend)

	// send funds from one of the double spends further
	spentInput := spend(ledgerState, doubleSpentOutputs1[1])

	// aggregate further sent funds with other reality
	outputOfAggregatedReality := spend(ledgerState, spentInput, doubleSpentOutputs2[0])

	// double spend further spend to elevate aggregated reality
	spend(ledgerState, doubleSpentOutputs1[1])

	// double spend funds of aggregated reality
	// spend(ledgerState, spentInput, doubleSpentOutputs2[0])

	// spend funds of conflict in aggregated reality further
	// lastOutputOfAggregatedReality := spend(ledgerState, outputOfAggregatedReality)

	// spend(ledgerState, lastOutputOfAggregatedReality, doubleSpentOutputs3[1])
	spend(ledgerState, spend(ledgerState, spend(ledgerState, outputOfAggregatedReality, spend(ledgerState, doubleSpentOutputs3[1]))))

	time.Sleep(1000 * time.Millisecond)

	objectstorage.WaitForWritesToFlush()

	_ = ledgerState.GenerateRealityVisualization("realities.png")
	_ = NewVisualizer(ledgerState).RenderTransferOutputs("outputs.png")
}

func TestElevate(t *testing.T) {
	ledgerState := NewLedgerState("testLedger").Prune().AddTransferOutput(
		transferHash1, addressHash1, coloredcoins.NewColoredBalance(eth, 1337), coloredcoins.NewColoredBalance(iota_, 1338),
	)

	// create first legit spend
	if err := ledgerState.BookTransfer(transfer.NewTransfer(transferHash2).AddInput(
		transfer.NewOutputReference(transferHash1, addressHash1),
	).AddOutput(
		addressHash2, coloredcoins.NewColoredBalance(iota_, 1338),
	).AddOutput(
		addressHash2, coloredcoins.NewColoredBalance(eth, 1337),
	)); err != nil {
		t.Error(err)
	}

	// send funds further
	if err := ledgerState.BookTransfer(transfer.NewTransfer(transferHash3).AddInput(
		transfer.NewOutputReference(transferHash2, addressHash2),
	).AddOutput(
		addressHash4, coloredcoins.NewColoredBalance(iota_, 1338),
	).AddOutput(
		addressHash4, coloredcoins.NewColoredBalance(eth, 1337),
	)); err != nil {
		t.Error(err)
	}

	if err := ledgerState.BookTransfer(transfer.NewTransfer(transferHash4).AddInput(
		transfer.NewOutputReference(transferHash2, addressHash2),
	).AddOutput(
		addressHash4, coloredcoins.NewColoredBalance(iota_, 1338),
	).AddOutput(
		addressHash4, coloredcoins.NewColoredBalance(eth, 1337),
	)); err != nil {
		t.Error(err)
	}

	// aggregate realities
	if err := ledgerState.BookTransfer(transfer.NewTransfer(transferHash6).AddInput(
		transfer.NewOutputReference(transferHash3, addressHash4),
	).AddInput(
		transfer.NewOutputReference(transferHash4, addressHash4),
	).AddOutput(
		addressHash6, coloredcoins.NewColoredBalance(iota_, 2676),
	).AddOutput(
		addressHash6, coloredcoins.NewColoredBalance(eth, 2674),
	)); err != nil {
		t.Error(err)
	}

	// create double spend for first transfer
	if err := ledgerState.BookTransfer(transfer.NewTransfer(transferHash5).AddInput(
		transfer.NewOutputReference(transferHash1, addressHash1),
	).AddOutput(
		addressHash5, coloredcoins.NewColoredBalance(iota_, 1338),
	).AddOutput(
		addressHash5, coloredcoins.NewColoredBalance(eth, 1337),
	)); err != nil {
		t.Error(err)
	}

	time.Sleep(1000 * time.Millisecond)

	objectstorage.WaitForWritesToFlush()

	ledgerState.ForEachTransferOutput(func(object *objectstorage.CachedObject) bool {
		object.Consume(func(object objectstorage.StorableObject) {
			fmt.Println(object.(*transfer.Output))
		})

		return true
	})
}
