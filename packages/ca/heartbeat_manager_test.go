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
	ownIdentity := identity.GenerateRandomIdentity()
	neighborIdentity := identity.GenerateRandomIdentity()

	// generate first heartbeat ////////////////////////////////////////////////////////////////////////////////////////

	heartbeatManager1 := NewHeartbeatManager(ownIdentity)
	heartbeatManager1.AddNeighbor(neighborIdentity)
	heartbeatManager1.InitialDislike(generateRandomTransactionId())
	heartbeatManager1.InitialDislike(generateRandomTransactionId())
	heartbeatManager1.InitialLike(generateRandomTransactionId())

	heartbeat1, err := heartbeatManager1.GenerateHeartbeat()
	if err != nil {
		t.Error(err)

		return
	}

	fmt.Println(heartbeat1)

	heartbeatManager2 := NewHeartbeatManager(neighborIdentity)
	heartbeatManager2.AddNeighbor(ownIdentity)
	err = heartbeatManager2.ApplyHeartbeat(heartbeat1)
	if err != nil {
		t.Error(err)

		return
	}

	heartbeat2, err := heartbeatManager2.GenerateHeartbeat()
	if err != nil {
		t.Error(err)

		return
	}

	fmt.Println(heartbeat2)
}
