package transaction

import (
	"fmt"
	"testing"

	"github.com/magiconair/properties/assert"

	"github.com/iotaledger/goshimmer/packages/binary/identity"
)

func TestNew(t *testing.T) {
	newTransaction1 := New(Id{}, Id{}, identity.Generate(), NewDataPayload([]byte("test")))
	assert.Equal(t, newTransaction1.VerifySignature(), true)

	newTransaction2 := New(newTransaction1.GetId(), Id{}, identity.Generate(), NewDataPayload([]byte("test1")))
	assert.Equal(t, newTransaction2.VerifySignature(), true)

	newTransaction3, _ := FromBytes(newTransaction2.GetBytes())
	assert.Equal(t, newTransaction3.VerifySignature(), true)

	fmt.Println(newTransaction1)
	fmt.Println(newTransaction2)
	fmt.Println(newTransaction3)
}
