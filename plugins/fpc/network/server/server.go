package server

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/iotaledger/goshimmer/packages/fpc"
	pb "github.com/iotaledger/goshimmer/plugins/fpc/network/query"
	"github.com/iotaledger/goshimmer/plugins/tangle"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/iota.go/trinary"
	"google.golang.org/grpc"
)

// queryServer defines the struct of an FPC query server
type queryServer struct {
	fpc *fpc.Instance
}

func newServer(fpc *fpc.Instance) *queryServer {
	return &queryServer{
		fpc: fpc,
	}
}

// GetOpinion returns the opinions of the given txs
func (s *queryServer) GetOpinion(ctx context.Context, req *pb.QueryRequest) (reply *pb.QueryReply, err error) {
	opinions := make([]fpc.Opinion, len(req.TxHash))
	for i, tx := range req.TxHash {
		opinions[i] = s.retrieveOpinion(tx)
	}
	reply = &pb.QueryReply{
		Opinion: opinions,
	}
	return reply, nil
}

func (s *queryServer) retrieveOpinion(tx fpc.ID) (opinion fpc.Opinion) {
	// temporary C until we have it properly defined somewhere
	const C = time.Second * 0

	//FPC lookup
	opinion, ok := s.fpc.GetInterimOpinion(tx)
	if ok {
		return opinion
	}
	//DB lookup
	txMetadata, err := tangle.GetTransactionMetadata(trinary.Trytes(tx))
	if err == nil && (txMetadata.GetFinalized() || txMetadata.GetReceivedTime().Add(C).Before(time.Now())) {
		return txMetadata.GetLiked()
	}

	return fpc.Dislike
}

// run starts a new server for replying to incoming queries
func run(address, port string, fpc *fpc.Instance) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", address+":"+port)
	if err != nil {
		return nil, err
	}
	var opts []grpc.ServerOption

	server := newServer(fpc)

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterFPCQueryServer(grpcServer, server)

	// Starts as a goroutine so that it does not block
	// and can return the grpcServer pointer to the caller
	go grpcServer.Serve(lis)
	return grpcServer, err
}

func RunServer(shutdownSignal <-chan struct{}, log *logger.Logger, fpc *fpc.Instance) {
	defer log.Info("Stopping FPC Server ... done")

	log.Info("Starting FPC Server (port " + strconv.Itoa(15666) + ") ...")

	log.Info("Starting FPC Server (port " + strconv.Itoa(15666) + ") ... done")

	server, _ := run("0.0.0.0", strconv.Itoa(15666), fpc)

	// Waits until receives a shutdown signal

	<-shutdownSignal
	log.Info("Stopping FPC Server ...")
	server.GracefulStop()

}
