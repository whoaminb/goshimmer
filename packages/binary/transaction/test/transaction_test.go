package test

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/iotaledger/goshimmer/packages/binary/async"

	"github.com/iotaledger/goshimmer/packages/binary/signature/ed25119"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/coloredcoins"

	"github.com/panjf2000/ants/v2"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"

	"github.com/iotaledger/goshimmer/packages/binary/transaction"

	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/identity"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/data"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/valuetransfer"
	"github.com/stretchr/testify/assert"
)

func BenchmarkVerifyDataTransactions(b *testing.B) {
	var pool async.WorkerPool
	pool.Tune(runtime.NumCPU() * 2)

	transactions := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		tx := transaction.New(transaction.EmptyId, transaction.EmptyId, identity.Generate(), data.New([]byte("some data")))

		if marshaledTransaction, err := tx.MarshalBinary(); err != nil {
			b.Error(err)
		} else {
			transactions[i] = marshaledTransaction
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		currentIndex := i
		pool.Submit(func() {
			if tx, err := transaction.FromBytes(transactions[currentIndex]); err != nil {
				b.Error(err)
			} else {
				tx.VerifySignature()
			}
		})
	}

	pool.Shutdown()
}

func BenchmarkVerifyValueTransactions(b *testing.B) {
	var pool async.WorkerPool
	pool.Tune(runtime.NumCPU() * 2)

	keyPairOfSourceAddress := ed25119.GenerateKeyPair()
	keyPairOfTargetAddress := ed25119.GenerateKeyPair()

	transactions := make([][]byte, b.N)
	for i := 0; i < b.N; i++ {
		tx := transaction.New(transaction.EmptyId, transaction.EmptyId, identity.Generate(), valuetransfer.New().
			AddInput(transfer.NewHash("test"), address.FromPublicKey(keyPairOfSourceAddress.PublicKey)).
			AddOutput(address.FromPublicKey(keyPairOfTargetAddress.PublicKey), coloredcoins.NewColoredBalance(coloredcoins.NewColor("IOTA"), 12)).
			Sign(keyPairOfSourceAddress))

		if marshaledTransaction, err := tx.MarshalBinary(); err != nil {
			b.Error(err)
		} else {
			transactions[i] = marshaledTransaction
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		currentIndex := i
		pool.Submit(func() {
			if tx, err := transaction.FromBytes(transactions[currentIndex]); err != nil {
				b.Error(err)
			} else {
				tx.VerifySignature()
				tx.GetPayload().(*valuetransfer.ValueTransfer).VerifySignatures()
			}
		})
	}

	pool.Shutdown()
}

func BenchmarkVerifySignature(b *testing.B) {
	pool, _ := ants.NewPool(80, ants.WithNonblocking(false))

	transactions := make([]*transaction.Transaction, b.N)
	for i := 0; i < b.N; i++ {
		transactions[i] = transaction.New(transaction.EmptyId, transaction.EmptyId, identity.Generate(), data.New([]byte("test")))
		transactions[i].GetBytes()
	}

	var wg sync.WaitGroup

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(1)

		currentIndex := i
		if err := pool.Submit(func() {
			transactions[currentIndex].VerifySignature()

			wg.Done()
		}); err != nil {
			b.Error(err)

			return
		}
	}

	wg.Wait()
}

func TestNew(t *testing.T) {
	newTransaction1 := transaction.New(transaction.EmptyId, transaction.EmptyId, identity.Generate(), data.New([]byte("test")))
	assert.Equal(t, newTransaction1.VerifySignature(), true)

	keyPairOfSourceAddress := ed25119.GenerateKeyPair()
	keyPairOfTargetAddress := ed25119.GenerateKeyPair()

	newValueTransaction1 := transaction.New(transaction.EmptyId, transaction.EmptyId, identity.Generate(),
		valuetransfer.New().
			AddInput(transfer.NewHash("test"), address.FromPublicKey(keyPairOfSourceAddress.PublicKey)).
			AddOutput(address.FromPublicKey(keyPairOfTargetAddress.PublicKey), coloredcoins.NewColoredBalance(coloredcoins.NewColor("IOTA"), 12)).
			Sign(keyPairOfSourceAddress),
	)
	assert.Equal(t, newValueTransaction1.VerifySignature(), true)

	newValueTransaction2, _ := transaction.FromBytes(newValueTransaction1.GetBytes())
	assert.Equal(t, newValueTransaction2.VerifySignature(), true)

	fmt.Println(newValueTransaction1.GetPayload().(*valuetransfer.ValueTransfer).MarshalBinary())
	fmt.Println(newValueTransaction2.GetPayload().(*valuetransfer.ValueTransfer).MarshalBinary())

	if newValueTransaction2.GetPayload().GetType() == valuetransfer.Type {
		fmt.Println(newValueTransaction1.GetPayload().(*valuetransfer.ValueTransfer).VerifySignatures())
		fmt.Println(newValueTransaction2.GetPayload().(*valuetransfer.ValueTransfer).VerifySignatures())
	}

	newTransaction2 := transaction.New(newTransaction1.GetId(), transaction.EmptyId, identity.Generate(), data.New([]byte("test1")))
	assert.Equal(t, newTransaction2.VerifySignature(), true)

	if newTransaction1.GetPayload().GetType() == data.Type {
		fmt.Println("DATA TRANSACTION")
	}

	newTransaction3, _ := transaction.FromBytes(newTransaction2.GetBytes())
	assert.Equal(t, newTransaction3.VerifySignature(), true)

	fmt.Println(newTransaction1)
	fmt.Println(newTransaction2)
	fmt.Println(newTransaction3)

	//fmt.Println(newValueTransaction1)
	//fmt.Println(newValueTransaction2)
}
