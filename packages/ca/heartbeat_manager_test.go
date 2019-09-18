package ca

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/iotaledger/goshimmer/packages/identity"
)

func generateRandomTransactionId() (result []byte) {
	result = make([]byte, 50)
	rand.Read(result)

	return
}

func TestHeartbeatManager_GenerateHeartbeat(t *testing.T) {
	transactionId1 := generateRandomTransactionId()
	transactionId2 := generateRandomTransactionId()

	heartbeatManager := NewHeartbeatManager(identity.GenerateRandomIdentity())
	heartbeatManager.InitialDislike(transactionId1)
	heartbeatManager.InitialDislike(transactionId2)

	heartbeatManager.InitialLike(generateRandomTransactionId())

	result, err := heartbeatManager.GenerateHeartbeat()
	if err != nil {
		t.Error(err)

		return
	}

	fmt.Println(result)

	result, err = heartbeatManager.GenerateHeartbeat()
	if err != nil {
		t.Error(err)

		return
	}

	fmt.Println(result)
}
