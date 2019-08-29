package heartbeat

import (
	"encoding/hex"
	"strconv"
	"sync"

	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/marshaling"

	"github.com/golang/protobuf/proto"
	heartbeatProto "github.com/iotaledger/goshimmer/packages/ca/heartbeat/proto"
)

type OpinionStatement struct {
	previousStatementHash []byte
	nodeId                string
	time                  uint64
	toggledTransactions   []*ToggledTransaction
	signature             []byte
	hash                  []byte

	previousStatementHashMutex sync.RWMutex
	nodeIdMutex                sync.RWMutex
	timeMutex                  sync.RWMutex
	toggledTransactionsMutex   sync.RWMutex
	signatureMutex             sync.RWMutex
	hashMutex                  sync.RWMutex
}

func NewOpinionStatement() *OpinionStatement {
	return &OpinionStatement{}
}

func (opinionStatement *OpinionStatement) GetPreviousStatementHash() []byte {
	opinionStatement.previousStatementHashMutex.RLock()
	defer opinionStatement.previousStatementHashMutex.RUnlock()

	return opinionStatement.previousStatementHash
}

func (opinionStatement *OpinionStatement) SetPreviousStatementHash(previousStatementHash []byte) {
	opinionStatement.previousStatementHashMutex.Lock()
	defer opinionStatement.previousStatementHashMutex.Unlock()

	opinionStatement.previousStatementHash = previousStatementHash
}

func (opinionStatement *OpinionStatement) GetNodeId() string {
	opinionStatement.nodeIdMutex.RLock()
	defer opinionStatement.nodeIdMutex.RUnlock()

	return opinionStatement.nodeId
}

func (opinionStatement *OpinionStatement) SetNodeId(nodeId string) {
	opinionStatement.nodeIdMutex.Lock()
	defer opinionStatement.nodeIdMutex.Unlock()

	opinionStatement.nodeId = nodeId
}

func (opinionStatement *OpinionStatement) GetTime() uint64 {
	opinionStatement.timeMutex.RLock()
	defer opinionStatement.timeMutex.RUnlock()

	return opinionStatement.time
}

func (opinionStatement *OpinionStatement) SetTime(time uint64) {
	opinionStatement.timeMutex.Lock()
	defer opinionStatement.timeMutex.Unlock()

	opinionStatement.time = time
}

func (opinionStatement *OpinionStatement) GetToggledTransactions() []*ToggledTransaction {
	opinionStatement.toggledTransactionsMutex.RLock()
	defer opinionStatement.toggledTransactionsMutex.RUnlock()

	return opinionStatement.toggledTransactions
}

func (opinionStatement *OpinionStatement) SetToggledTransactions(toggledTransactions []*ToggledTransaction) {
	opinionStatement.toggledTransactionsMutex.Lock()
	defer opinionStatement.toggledTransactionsMutex.Unlock()

	opinionStatement.toggledTransactions = toggledTransactions
}

func (opinionStatement *OpinionStatement) GetSignature() []byte {
	opinionStatement.signatureMutex.RLock()
	defer opinionStatement.signatureMutex.RUnlock()

	return opinionStatement.signature
}

func (opinionStatement *OpinionStatement) SetSignature(signature []byte) {
	opinionStatement.signatureMutex.Lock()
	defer opinionStatement.signatureMutex.Unlock()

	opinionStatement.signature = signature
}

func (opinionStatement *OpinionStatement) GetHash() []byte {
	opinionStatement.hashMutex.RLock()
	defer opinionStatement.hashMutex.RUnlock()

	return opinionStatement.hash
}

func (opinionStatement *OpinionStatement) SetHash(hash []byte) {
	opinionStatement.hashMutex.Lock()
	defer opinionStatement.hashMutex.Unlock()

	opinionStatement.hash = hash
}

func (opinionStatement *OpinionStatement) FromProto(proto proto.Message) {
	protoOpinionStatement := proto.(*heartbeatProto.OpinionStatement)

	opinionStatement.previousStatementHash = protoOpinionStatement.PreviousStatementHash
	opinionStatement.nodeId = protoOpinionStatement.NodeId
	opinionStatement.time = protoOpinionStatement.Time
	opinionStatement.signature = protoOpinionStatement.Signature

	opinionStatement.toggledTransactions = make([]*ToggledTransaction, len(protoOpinionStatement.ToggledTransactions))
	for i, toggledTransaction := range protoOpinionStatement.ToggledTransactions {
		var newToggledTransaction ToggledTransaction
		newToggledTransaction.FromProto(toggledTransaction)

		opinionStatement.toggledTransactions[i] = &newToggledTransaction
	}
}

func (opinionStatement *OpinionStatement) ToProto() proto.Message {
	toggledTransactions := make([]*heartbeatProto.ToggledTransaction, len(opinionStatement.toggledTransactions))
	for i, toggledTransaction := range opinionStatement.toggledTransactions {
		toggledTransactions[i] = toggledTransaction.ToProto().(*heartbeatProto.ToggledTransaction)
	}

	return &heartbeatProto.OpinionStatement{
		PreviousStatementHash: opinionStatement.previousStatementHash,
		NodeId:                opinionStatement.nodeId,
		Time:                  opinionStatement.time,
		ToggledTransactions:   toggledTransactions,
		Signature:             opinionStatement.signature,
	}
}

func (opinionStatement *OpinionStatement) MarshalBinary() ([]byte, errors.IdentifiableError) {
	return marshaling.Marshal(opinionStatement)
}

func (opinionStatement *OpinionStatement) UnmarshalBinary(data []byte) (err errors.IdentifiableError) {
	return marshaling.Unmarshal(opinionStatement, data, &heartbeatProto.OpinionStatement{})
}

func (opinionStatement *OpinionStatement) String() string {
	return "OpinionStatement {\n" +
		"    previousStatementHash: 0x" + hex.EncodeToString(opinionStatement.previousStatementHash) + "\n" +
		"    nodeId:                " + opinionStatement.nodeId + "\n" +
		"    time:                  " + strconv.Itoa(int(opinionStatement.time)) + "\n" +
		"    toggledTransactions:   [" + "" + "]\n" +
		"    signature:             0x" + hex.EncodeToString(opinionStatement.signature) + "\n" +
		"}"
}
