package ca

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/typeutils"

	"github.com/iotaledger/goshimmer/packages/events"

	"github.com/iotaledger/goshimmer/packages/identity"
)

func generateRandomTransactionId() (result []byte) {
	result = make([]byte, 50)
	rand.Read(result)

	return
}

type virtualNode struct {
	identity         *identity.Identity
	heartbeatManager *HeartbeatManager
}

func generateVirtualNetwork(numberOfNodes int) (result []*virtualNode) {
	for i := 0; i < numberOfNodes; i++ {
		nodeIdentity := identity.GenerateRandomIdentity()

		virtualNode := &virtualNode{
			identity:         nodeIdentity,
			heartbeatManager: NewHeartbeatManager(nodeIdentity),
		}

		result = append(result, virtualNode)
	}

	for i := 0; i < numberOfNodes; i++ {
		for j := 0; j < numberOfNodes; j++ {
			if i != j {
				result[i].heartbeatManager.AddNeighbor(result[j].identity)
			}
		}
	}

	go func() {
		for {
			for _, node := range result {
				heartbeat, err := node.heartbeatManager.GenerateHeartbeat()
				if err != nil {
					fmt.Println(err)

					return
				}

				for _, otherNode := range result {
					if otherNode != node {
						otherNode.heartbeatManager.ApplyHeartbeat(heartbeat)
					}
				}
			}

			time.Sleep(700 * time.Millisecond)
		}
	}()

	return
}

func TestConsensus(t *testing.T) {
	virtualNetwork := generateVirtualNetwork(5)

	transactionId := generateRandomTransactionId()

	virtualNetwork[0].heartbeatManager.InitialDislike(transactionId)
	virtualNetwork[1].heartbeatManager.InitialDislike(transactionId)
	virtualNetwork[2].heartbeatManager.InitialDislike(transactionId)

	time.Sleep(1 * time.Second)

	fmt.Println(virtualNetwork[4].heartbeatManager.opinions.GetOpinion(typeutils.BytesToString(transactionId)).IsLiked())
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

	fmt.Println(heartbeat2)

	if err = heartbeatManager1.ApplyHeartbeat(heartbeat2); err != nil {
		t.Error(err)

		return
	}

	heartbeat3, err := heartbeatManager1.GenerateHeartbeat()
	if err != nil {
		t.Error(err)

		return
	}

	fmt.Println(heartbeat3)
}
