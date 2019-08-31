package heartbeat

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/stringify"

	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/marshaling"

	"github.com/golang/protobuf/proto"
	heartbeatProto "github.com/iotaledger/goshimmer/packages/ca/heartbeat/proto"
)

type ToggledTransaction struct {
	transactionId    []byte
	initialStatement bool
	finalStatement   bool

	transactionIdMutex    sync.RWMutex
	initialStatementMutex sync.RWMutex
	finalStatementMutex   sync.RWMutex
}

func NewToggledTransaction() *ToggledTransaction {
	return &ToggledTransaction{}
}

func (toggledTransaction *ToggledTransaction) GetTransactionId() []byte {
	toggledTransaction.transactionIdMutex.RLock()
	defer toggledTransaction.transactionIdMutex.RUnlock()

	return toggledTransaction.transactionId
}

func (toggledTransaction *ToggledTransaction) SetTransactionId(transactionId []byte) {
	toggledTransaction.transactionIdMutex.Lock()
	defer toggledTransaction.transactionIdMutex.Unlock()

	toggledTransaction.transactionId = transactionId
}

func (toggledTransaction *ToggledTransaction) IsInitialStatement() bool {
	toggledTransaction.initialStatementMutex.RLock()
	defer toggledTransaction.initialStatementMutex.RUnlock()

	return toggledTransaction.initialStatement
}

func (toggledTransaction *ToggledTransaction) SetInitialStatement(initialStatement bool) {
	toggledTransaction.initialStatementMutex.Lock()
	defer toggledTransaction.initialStatementMutex.Unlock()

	toggledTransaction.initialStatement = initialStatement
}

func (toggledTransaction *ToggledTransaction) IsFinalStatement() bool {
	toggledTransaction.finalStatementMutex.RLock()
	defer toggledTransaction.finalStatementMutex.RUnlock()

	return toggledTransaction.finalStatement
}

func (toggledTransaction *ToggledTransaction) SetFinalStatement(finalStatement bool) {
	toggledTransaction.finalStatementMutex.Lock()
	defer toggledTransaction.finalStatementMutex.Unlock()

	toggledTransaction.finalStatement = finalStatement
}

func (toggledTransaction *ToggledTransaction) FromProto(proto proto.Message) {
	protoToggledTransaction := proto.(*heartbeatProto.ToggledTransaction)

	toggledTransaction.transactionId = protoToggledTransaction.TransactionId
	toggledTransaction.initialStatement = protoToggledTransaction.InitialStatement
	toggledTransaction.finalStatement = protoToggledTransaction.FinalStatement
}

func (toggledTransaction *ToggledTransaction) ToProto() proto.Message {
	return &heartbeatProto.ToggledTransaction{
		TransactionId:    toggledTransaction.transactionId,
		InitialStatement: toggledTransaction.initialStatement,
		FinalStatement:   toggledTransaction.finalStatement,
	}
}

func (toggledTransaction *ToggledTransaction) MarshalBinary() ([]byte, errors.IdentifiableError) {
	return marshaling.Marshal(toggledTransaction)
}

func (toggledTransaction *ToggledTransaction) UnmarshalBinary(data []byte) (err errors.IdentifiableError) {
	return marshaling.Unmarshal(toggledTransaction, data, &heartbeatProto.ToggledTransaction{})
}

func (toggledTransaction *ToggledTransaction) String() string {
	return stringify.Struct("ToggledTransaction",
		stringify.StructField("transactionId", toggledTransaction.transactionId),
		stringify.StructField("initialStatement", toggledTransaction.initialStatement),
		stringify.StructField("finalStatement", toggledTransaction.finalStatement),
	)
}
