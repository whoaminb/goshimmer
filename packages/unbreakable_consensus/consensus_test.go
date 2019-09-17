package unbreakable_consensus

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/unbreakable_consensus/social_consensus"
)

func generateElders(totalMana int, nodeCount int) []*social_consensus.Node {
	result := make([]*social_consensus.Node, nodeCount)

	for i := 0; i < nodeCount; i++ {
		newNode := social_consensus.NewNode("elder" + strconv.Itoa(i))
		newNode.SetReputation(uint64(totalMana / nodeCount))

		result[i] = newNode
	}

	return result
}

func TestConsensus(t *testing.T) {
	rand.Seed(time.Now().Unix())

	elders := generateElders(100, 10)

	society := social_consensus.NewSociety(elders)

	fmt.Println(society.GetReferencedElderReputation(social_consensus.ElderMask(1).Union(social_consensus.ElderMask(1))))

	fmt.Println(society.GetRandomElder().GetElderMask())

	/*

		selectRandomElder := func() elder.Elder {
			return nodes[rand.Intn(len(nodes))]
		}

		epochRegister := NewEpochRegister()

		genesis := NewTransaction("GENESIS")

		//selectRandomNode().Attach(genesis)

		fmt.Println(selectRandomElder())

		fmt.Println(genesis)

		fmt.Println(epochRegister.GetEpoch(0))
	*/
}
