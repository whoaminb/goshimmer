package ca

import (
	"bytes"

	"github.com/iotaledger/goshimmer/packages/typeutils"

	"github.com/iotaledger/goshimmer/packages/events"

	"github.com/iotaledger/goshimmer/packages/ca/heartbeat"
	"github.com/iotaledger/goshimmer/packages/errors"
)

type StatementChain struct {
	Events StatementChainEvents

	statements            map[string]*heartbeat.OpinionStatement
	missingStatements     map[string]bool
	idleCounter           int
	lastAppliedStatement  *heartbeat.OpinionStatement
	lastReceivedStatement *heartbeat.OpinionStatement
}

func NewStatementChain() *StatementChain {
	return &StatementChain{
		Events: StatementChainEvents{
			Reset:            events.NewEvent(events.CallbackCaller),
			StatementMissing: events.NewEvent(HashCaller),
		},
		statements:        make(map[string]*heartbeat.OpinionStatement),
		missingStatements: make(map[string]bool),
	}
}

func (statementChain *StatementChain) IncreaseIdleCounter() {
	statementChain.idleCounter++
}

func (statementChain *StatementChain) ResetIdleCounter() {
	statementChain.idleCounter = 0
}

func (statementChain *StatementChain) GetIdleCounter() int {
	return statementChain.idleCounter
}

func (statementChain *StatementChain) AddStatement(statement *heartbeat.OpinionStatement) bool {
	previousStatementHash := statement.GetPreviousStatementHash()
	lastAppliedMainStatement := statementChain.lastAppliedStatement

	if len(previousStatementHash) == 0 {
		statementChain.Reset()
	} else if lastAppliedMainStatement != nil && !bytes.Equal(lastAppliedMainStatement.GetHash(), previousStatementHash) {
		if (statement.GetTime() - lastAppliedMainStatement.GetTime()) >= MAX_STATEMENT_TIMEOUT {
			statementChain.Reset()
		} else if !statementChain.StatementExists(previousStatementHash) {
			statementChain.addMissingStatement(previousStatementHash)
			statementChain.addStatement(statement)

			return false
		}
	}

	statementChain.addStatement(statement)

	return true
}

func (statementChain *StatementChain) addStatement(statement *heartbeat.OpinionStatement) {
	statementChain.statements[typeutils.BytesToString(statement.GetHash())] = statement
}

func (statementChain *StatementChain) addMissingStatement(statementHash []byte) {
	statementChain.missingStatements[typeutils.BytesToString(statementHash)] = true

	statementChain.Events.StatementMissing.Trigger(statementHash)
}

func (statementChain *StatementChain) GetStatement(statementHash []byte) *heartbeat.OpinionStatement {
	return statementChain.statements[typeutils.BytesToString(statementHash)]
}

func (statementChain *StatementChain) Reset() {
	statementChain.Events.Reset.Trigger()
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
