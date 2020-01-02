package tangle

import (
	"fmt"
	"testing"

	"github.com/iotaledger/goshimmer/packages/binary/identity"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/data"
)

func BenchmarkTangle_AttachTransaction(b *testing.B) {
	tangle := New([]byte("TEST_BINARY_TANGLE"))
	if err := tangle.Prune(); err != nil {
		b.Error(err)

		return
	}

	testIdentity := identity.Generate()

	transactionBytes := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		transactionBytes[i] = transaction.New(transaction.EmptyId, transaction.EmptyId, testIdentity, data.New([]byte("some data"))).GetBytes()
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if txToAttach, err := transaction.FromBytes(transactionBytes[i]); err != nil {
			b.Error(err)

			return
		} else {
			tangle.AttachTransaction(txToAttach)
		}
	}

	tangle.Shutdown()
}

func TestTangle_AttachTransaction(t *testing.T) {
	tangle := New([]byte("TEST_BINARY_TANGLE"))
	if err := tangle.Prune(); err != nil {
		t.Error(err)

		return
	}

	newTransaction1 := transaction.New(transaction.EmptyId, transaction.EmptyId, identity.Generate(), data.New([]byte("some data")))
	newTransaction2 := transaction.New(newTransaction1.GetId(), transaction.EmptyId, identity.Generate(), data.New([]byte("some other data")))

	fmt.Println("ATTACH", newTransaction2.GetId())
	tangle.AttachTransaction(newTransaction2)

	fmt.Println("ATTACH", newTransaction1.GetId())
	tangle.AttachTransaction(newTransaction1)

	tangle.Shutdown()
}
