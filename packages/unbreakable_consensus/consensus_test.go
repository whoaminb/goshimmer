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

func issueTransaction() {

}

func BenchmarkTPS(b *testing.B) {
	rand.Seed(time.Now().Unix())

	elders := generateElders(10000, 20)

	society := social_consensus.NewSociety(elders)
	currentTx := social_consensus.NewTransaction()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tx := social_consensus.NewTransaction()
		tx.SetIssuer(society.GetRandomElder())
		tx.SetElder(society.GetRandomElder())
		tx.Attach(currentTx, currentTx)

		currentTx = tx
	}
}

func TestConsensus(t *testing.T) {
	rand.Seed(time.Now().Unix())

	elders := generateElders(100, 10)

	society := social_consensus.NewSociety(elders)

	genesis := social_consensus.NewTransaction()

	tx1 := social_consensus.NewTransaction()
	tx1.SetIssuer(society.GetRandomElder())
	tx1.SetElder(society.GetRandomElder())
	tx1.Attach(genesis, genesis)

	tx2 := social_consensus.NewTransaction()
	tx2.SetIssuer(society.GetRandomElder())
	tx2.SetElder(society.GetRandomElder())
	tx2.Attach(tx1, genesis)

	fmt.Println(tx2)

	fmt.Println(society.GetRandomElder())

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
