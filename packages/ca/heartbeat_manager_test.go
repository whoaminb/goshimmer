package ca

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/iotaledger/goshimmer/packages/identity"
)

func TestHeartbeatManager_GenerateHeartbeat(t *testing.T) {
	transactionId1 := make([]byte, 50)
	rand.Read(transactionId1)

	heartbeatManager := NewHeartbeatManager(identity.GenerateRandomIdentity())
	heartbeatManager.SetInitialOpinion(transactionId1)

	result, err := heartbeatManager.GenerateHeartbeat()
	if err != nil {
		t.Error(err)

		return
	}

	fmt.Println(result.GetMainStatement())
}
