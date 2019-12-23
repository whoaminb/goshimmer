package test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/panjf2000/ants/v2"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"

	"github.com/iotaledger/goshimmer/packages/binary/transaction"

	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/identity"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/data"
	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/valuetransfer"
	"github.com/stretchr/testify/assert"
)

func BenchmarkVerifySignature(b *testing.B) {
	transaction := transaction.New(transaction.EmptyId, transaction.EmptyId, identity.Generate(), data.New([]byte("test")))

	var wg sync.WaitGroup

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(1)

		ants.Submit(func() {
			transaction.VerifySignature()

			wg.Done()
		})
	}

	wg.Wait()
}

func TestNew(t *testing.T) {
	newTransaction1 := transaction.New(transaction.EmptyId, transaction.EmptyId, identity.Generate(), data.New([]byte("test")))
	assert.Equal(t, newTransaction1.VerifySignature(), true)

	valueTransfer := valuetransfer.New().AddInput(transfer.NewHash("test"), address.Random())

	newValueTransaction1 := transaction.New(transaction.EmptyId, transaction.EmptyId, identity.Generate(), valueTransfer)
	assert.Equal(t, newValueTransaction1.VerifySignature(), true)

	newValueTransaction2, _ := transaction.FromBytes(newValueTransaction1.GetBytes())
	assert.Equal(t, newValueTransaction2.VerifySignature(), true)

	if newValueTransaction1.GetPayload().GetType() == valuetransfer.Type {
		fmt.Println("VALUE TRANSFER TRANSACTION")
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
