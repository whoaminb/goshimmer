package ca

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/iotaledger/goshimmer/packages/events"

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
	droppedNeighborIdentity := identity.GenerateRandomIdentity()

	// generate first heartbeat ////////////////////////////////////////////////////////////////////////////////////////

	heartbeatManager1 := NewHeartbeatManager(ownIdentity)
	heartbeatManager1.AddNeighbor(neighborIdentity)
	heartbeatManager1.AddNeighbor(droppedNeighborIdentity)
	heartbeatManager1.RemoveNeighbor(droppedNeighborIdentity)
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

	heartbeatManager2.Events.AddNeighbor.Attach(events.NewClosure(func(neighborIdentity *identity.Identity, neighborManager *NeighborManager) {
		neighborManager.Events.ChainReset.Attach(events.NewClosure(func() {
			fmt.Println("RESET")
		}))
	}))
	heartbeatManager2.Events.RemoveNeighbor.Attach(events.NewClosure(func(neighborIdentity *identity.Identity, neighborManager *NeighborManager) {
		fmt.Println("Rest")
	}))

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

	if err = heartbeatManager1.ApplyHeartbeat(heartbeat2); err != nil {
		t.Error(err)

		return
	}

	fmt.Println(heartbeat2)

	heartbeat3, err := heartbeatManager1.GenerateHeartbeat()
	if err != nil {
		t.Error(err)

		return
	}

	fmt.Println(heartbeat3)
}
