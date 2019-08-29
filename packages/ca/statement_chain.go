package ca

import (
	"github.com/iotaledger/goshimmer/packages/ca/heartbeat"
	"github.com/iotaledger/goshimmer/packages/errors"
)

type StatementChain struct {
	lastAppliedStatement  *heartbeat.OpinionStatement
	lastReceivedStatement *heartbeat.OpinionStatement
}

func NewStatementChain() *StatementChain {
	return &StatementChain{}
}

func (statementChain *StatementChain) AddStatement(statement *heartbeat.OpinionStatement) {

}

func (statementChain *StatementChain) Reset() {

}

func (statementChain *StatementChain) Apply(statement *heartbeat.OpinionStatement) (err errors.IdentifiableError) {
	return
}

func (statementChain *StatementChain) SetLastReceivedStatement(statement *heartbeat.OpinionStatement) {
	statementChain.lastReceivedStatement = statement
}

func (statementChain *StatementChain) GetLastAppliedStatement() *heartbeat.OpinionStatement {
	return statementChain.lastAppliedStatement
}

func (statementChain *StatementChain) StatementExists(statementHash []byte) bool {
	return true
}
