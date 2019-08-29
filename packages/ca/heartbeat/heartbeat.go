package heartbeat

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/marshaling"

	"github.com/golang/protobuf/proto"
	heartbeatProto "github.com/iotaledger/goshimmer/packages/ca/heartbeat/proto"
)

type Heartbeat struct {
	nodeId             string
	mainStatement      *OpinionStatement
	neighborStatements map[string]*OpinionStatement
	signature          []byte

	nodeIdMutex             sync.RWMutex
	mainStatementMutex      sync.RWMutex
	neighborStatementsMutex sync.RWMutex
	signatureMutex          sync.RWMutex
}

func NewHeartbeat() *Heartbeat {
	return &Heartbeat{}
}

func (heartbeat *Heartbeat) GetNodeId() string {
	heartbeat.nodeIdMutex.RLock()
	defer heartbeat.nodeIdMutex.RLock()

	return heartbeat.nodeId
}

func (heartbeat *Heartbeat) SetNodeId(nodeId string) {
	heartbeat.nodeIdMutex.Lock()
	defer heartbeat.nodeIdMutex.Unlock()

	heartbeat.nodeId = nodeId
}

func (heartbeat *Heartbeat) GetMainStatement() *OpinionStatement {
	heartbeat.mainStatementMutex.RLock()
	defer heartbeat.mainStatementMutex.RUnlock()

	return heartbeat.mainStatement
}

func (heartbeat *Heartbeat) SetMainStatement(mainStatement *OpinionStatement) {
	heartbeat.mainStatementMutex.Lock()
	defer heartbeat.mainStatementMutex.Unlock()

	heartbeat.mainStatement = mainStatement
}

func (heartbeat *Heartbeat) GetNeighborStatements() map[string]*OpinionStatement {
	heartbeat.neighborStatementsMutex.RLock()
	defer heartbeat.neighborStatementsMutex.RUnlock()

	return heartbeat.neighborStatements
}

func (heartbeat *Heartbeat) SetNeighborStatements(neighborStatements map[string]*OpinionStatement) {
	heartbeat.neighborStatementsMutex.Lock()
	defer heartbeat.neighborStatementsMutex.Unlock()

	heartbeat.neighborStatements = neighborStatements
}

func (heartbeat *Heartbeat) GetSignature() []byte {
	heartbeat.signatureMutex.RLock()
	defer heartbeat.signatureMutex.RUnlock()

	return heartbeat.signature
}

func (heartbeat *Heartbeat) SetSignature(signature []byte) {
	heartbeat.signatureMutex.Lock()
	defer heartbeat.signatureMutex.Unlock()

	heartbeat.signature = signature
}

func (heartbeat *Heartbeat) FromProto(proto proto.Message) {
	protoHeartbeat := proto.(*heartbeatProto.HeartBeat)

	var mainStatement OpinionStatement
	mainStatement.FromProto(protoHeartbeat.MainStatement)

	neighborStatements := make(map[string]*OpinionStatement, len(protoHeartbeat.NeighborStatements))
	for _, neighborStatement := range protoHeartbeat.NeighborStatements {
		var newNeighborStatement OpinionStatement
		newNeighborStatement.FromProto(neighborStatement)

		neighborStatements[neighborStatement.NodeId] = &newNeighborStatement
	}

	heartbeat.nodeId = protoHeartbeat.NodeId
	heartbeat.mainStatement = &mainStatement
	heartbeat.neighborStatements = neighborStatements
	heartbeat.signature = protoHeartbeat.Signature
}

func (heartbeat *Heartbeat) ToProto() proto.Message {
	neighborStatements := make([]*heartbeatProto.OpinionStatement, len(heartbeat.neighborStatements))
	i := 0
	for _, neighborStatement := range heartbeat.neighborStatements {
		neighborStatements[i] = neighborStatement.ToProto().(*heartbeatProto.OpinionStatement)

		i++
	}

	return &heartbeatProto.HeartBeat{
		NodeId:             heartbeat.nodeId,
		MainStatement:      heartbeat.mainStatement.ToProto().(*heartbeatProto.OpinionStatement),
		NeighborStatements: neighborStatements,
		Signature:          heartbeat.signature,
	}
}

func (heartbeat *Heartbeat) MarshalBinary() ([]byte, errors.IdentifiableError) {
	return marshaling.Marshal(heartbeat)
}

func (heartbeat *Heartbeat) UnmarshalBinary(data []byte) (err errors.IdentifiableError) {
	return marshaling.Unmarshal(heartbeat, data, &heartbeatProto.HeartBeat{})
}
