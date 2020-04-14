package tests

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/plugins/webapi/autopeering"
	"github.com/iotaledger/goshimmer/tools/integration-tests/tester/framework"
	"github.com/stretchr/testify/require"
)

func TestNeighbors(t *testing.T) {
	for {
		for _, p := range f.Peers() {
			// get gossip neighbors
			resp, err := p.GetGossipNeighbors()
			require.NoError(t, err)
			gossipNeighbors := resp.Neighbors

			var gossipNeighborsClean []autopeering.Neighbor
			for _, n := range gossipNeighbors {
				if len(n.ID) == 0 {
					continue
				}
				gossipNeighborsClean = append(gossipNeighborsClean, n)
			}
			sort.Slice(gossipNeighborsClean, func(i, j int) bool {
				return strings.Compare(gossipNeighborsClean[i].ID, gossipNeighborsClean[j].ID) < 0
			})

			// get autopeering neighbors
			resp2, err := p.GetNeighbors(false)
			require.NoError(t, err)

			var autopeeringNeighbors []autopeering.Neighbor
			autopeeringNeighbors = append(autopeeringNeighbors, resp2.Accepted...)
			autopeeringNeighbors = append(autopeeringNeighbors, resp2.Chosen...)
			sort.Slice(autopeeringNeighbors, func(i, j int) bool {
				return strings.Compare(autopeeringNeighbors[i].ID, autopeeringNeighbors[j].ID) < 0
			})

			if len(gossipNeighborsClean) != len(autopeeringNeighbors) {
				printPeerWithNeighbors(p, gossipNeighborsClean, autopeeringNeighbors)
			}
		}

		fmt.Println("Waiting 10 secs...")
		time.Sleep(10 * time.Second)
	}
}

func printPeerWithNeighbors(peer *framework.Peer, gossipNeighbors, autopeeringNeighbors []autopeering.Neighbor) {
	fmt.Printf("-----------------------------\n")

	fmt.Println(peer.String())

	fmt.Printf("Gossip: %d\n", len(gossipNeighbors))
	for i, n := range gossipNeighbors {
		fmt.Printf("%d: %s\n", i, n.String())
	}

	fmt.Printf("Autopeering: %d\n", len(autopeeringNeighbors))
	for i, n := range autopeeringNeighbors {
		fmt.Printf("%d: %s\n", i, n.String())
	}

	fmt.Printf("-----------------------------\n\n")
}
