package network

import (
	"context"
	"log"
	"time"

	"github.com/iotaledger/goshimmer/packages/fpc"
	pb "github.com/iotaledger/goshimmer/plugins/fpc/network/query"
	"google.golang.org/grpc"
)

const (
	// TIMEOUT is the connection timeout
	TIMEOUT = 500 * time.Millisecond
)

// queryNode is the internal
func queryNode(txHash []fpc.ID, client pb.FPCQueryClient) (output []fpc.Opinion) {
	ctx, cancel := context.WithTimeout(context.Background(), TIMEOUT)
	defer cancel()

	// Prepare query
	query := &pb.QueryRequest{
		TxHash: txHash,
	}

	opinions, err := client.GetOpinion(ctx, query)
	if err != nil {
		log.Printf("%v.GetOpinion(_) = _, %v: \n", client, err)
		return output
	}

	return opinions.GetOpinion()
}

// QueryNode sends a query to a node and returns a list of opinions
func QueryNode(txHash []fpc.ID, nodeAddress string) (opinions []fpc.Opinion) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	// Connect to the node server
	conn, err := grpc.Dial(nodeAddress, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	// Setup a new client over the previous connection
	client := pb.NewFPCQueryClient(conn)

	// Send query
	opinions = queryNode(txHash, client)

	return opinions
}
