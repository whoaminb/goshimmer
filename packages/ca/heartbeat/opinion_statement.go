package heartbeat

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/identity"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/goshimmer/packages/stringify"

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

func (opinionStatement *OpinionStatement) Sign(identity *identity.Identity) (err errors.IdentifiableError) {
	if marshaledStatement, marshalErr := opinionStatement.MarshalBinary(); marshalErr == nil {
		if signature, signingErr := identity.Sign(marshaledStatement); signingErr == nil {
			opinionStatement.SetSignature(signature)
		} else {
			err = ErrSigningFailed.Derive(signingErr, "failed to sign opinion statement")
		}
	} else {
		err = marshalErr
	}

	return
}

func (opinionStatement *OpinionStatement) GetHash() []byte {
	opinionStatement.hashMutex.RLock()
	defer opinionStatement.hashMutex.RUnlock()

	if opinionStatement.hash == nil {
		marshaledStatement, marshalErr := opinionStatement.MarshalBinary()
		if marshalErr != nil {
			panic(marshalErr)
		}

		hash := blake2b.Sum256(marshaledStatement)

		opinionStatement.hash = hash[:]
	}

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
	return stringify.Struct("OpinionStatement",
		stringify.StructField("previousStatementHash", opinionStatement.previousStatementHash),
		stringify.StructField("nodeId", opinionStatement.nodeId),
		stringify.StructField("time", opinionStatement.time),
		stringify.StructField("toggledTransactions", opinionStatement.toggledTransactions),
		stringify.StructField("signature", opinionStatement.signature),
	)
}
