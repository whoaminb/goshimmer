package tangle

import (
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/events"

	"github.com/iotaledger/goshimmer/packages/binary/identity"
	"github.com/iotaledger/goshimmer/packages/binary/tangle/model/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/tangle/model/transaction/payload/data"
)

func BenchmarkTangle_AttachTransaction(b *testing.B) {
	tangle := New([]byte("TEST_BINARY_TANGLE"))
	if err := tangle.Prune(); err != nil {
		b.Error(err)

		return
	}

	testIdentity := identity.Generate()

	transactionBytes := make([]*transaction.Transaction, b.N)
	for i := 0; i < b.N; i++ {
		transactionBytes[i] = transaction.New(transaction.EmptyId, transaction.EmptyId, testIdentity, data.New([]byte("some data")))
		transactionBytes[i].GetBytes()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tangle.AttachTransaction(transactionBytes[i])
	}

	tangle.Shutdown()
}

func TestTangle_AttachTransaction(t *testing.T) {
	tangle := New([]byte("TEST_BINARY_TANGLE"))
	if err := tangle.Prune(); err != nil {
		t.Error(err)

		return
	}

	tangle.Events.TransactionMissing.Attach(events.NewClosure(func(transactionId transaction.Id) {
		fmt.Println("MISSING:", transactionId)
	}))

	tangle.Events.TransactionRemoved.Attach(events.NewClosure(func(transactionId transaction.Id) {
		fmt.Println("REMOVED:", transactionId)
	}))

	newTransaction1 := transaction.New(transaction.EmptyId, transaction.EmptyId, identity.Generate(), data.New([]byte("some data")))
	newTransaction2 := transaction.New(newTransaction1.GetId(), newTransaction1.GetId(), identity.Generate(), data.New([]byte("some other data")))

	fmt.Println("ATTACH", newTransaction2.GetId())
	tangle.AttachTransaction(newTransaction2)

	time.Sleep(37 * time.Second)

	fmt.Println("ATTACH", newTransaction1.GetId())
	tangle.AttachTransaction(newTransaction1)

	tangle.Shutdown()
}
