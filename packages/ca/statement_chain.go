package ca

import (
	"bytes"

	"github.com/iotaledger/goshimmer/packages/errors"

	"github.com/iotaledger/goshimmer/packages/typeutils"

	"github.com/iotaledger/goshimmer/packages/events"

	"github.com/iotaledger/goshimmer/packages/ca/heartbeat"
)

type StatementChain struct {
	Events StatementChainEvents

	pendingTransactionStatuses map[string]bool
	transactionStatuses        map[string]bool
	statements                 map[string]*heartbeat.OpinionStatement
	tail                       *heartbeat.OpinionStatement
}

func NewStatementChain() *StatementChain {
	return &StatementChain{
		Events: StatementChainEvents{
			Reset: events.NewEvent(events.CallbackCaller),
		},
		pendingTransactionStatuses: make(map[string]bool),
		transactionStatuses:        make(map[string]bool),
		statements:                 make(map[string]*heartbeat.OpinionStatement),
	}
}

func (statementChain *StatementChain) getTransactionStatus(transactionId string) (result bool, exists bool) {
	if result, exists = statementChain.pendingTransactionStatuses[transactionId]; exists {
		return
	}

	result, exists = statementChain.transactionStatuses[transactionId]

	return
}

func (statementChain *StatementChain) AddStatement(statement *heartbeat.OpinionStatement) errors.IdentifiableError {
	previousStatementHash := statement.GetPreviousStatementHash()
	lastAppliedStatement := statementChain.tail

	if len(previousStatementHash) == 0 || lastAppliedStatement != nil && !bytes.Equal(lastAppliedStatement.GetHash(), previousStatementHash) {
		statementChain.Reset()
	}

	for _, toggledTransaction := range statement.GetToggledTransactions() {
		transactionId := typeutils.BytesToString(toggledTransaction.GetTransactionId())

		if toggledTransaction.IsInitialStatement() {
			if _, exists := statementChain.getTransactionStatus(transactionId); exists {
				return ErrMalformedHeartbeat.Derive("two initial statements for the same transaction")
			}

			statementChain.pendingTransactionStatuses[transactionId] = false
		} else if toggledTransaction.IsFinalStatement() {
			// finalize -> clean up
		} else {
			if currentValue, exists := statementChain.getTransactionStatus(transactionId); exists {
				statementChain.pendingTransactionStatuses[transactionId] = !currentValue
			}
		}
	}

	statementChain.statements[typeutils.BytesToString(statement.GetHash())] = statement
	statementChain.tail = statement

	return nil
}

func (statementChain *StatementChain) ApplyPendingTransactionStatusChanges() {
	for transactionId, value := range statementChain.pendingTransactionStatuses {
		statementChain.transactionStatuses[transactionId] = value
	}

	statementChain.pendingTransactionStatuses = make(map[string]bool)
}

func (statementChain *StatementChain) GetStatement(statementHash []byte) *heartbeat.OpinionStatement {
	return statementChain.statements[typeutils.BytesToString(statementHash)]
}

func (statementChain *StatementChain) Reset() {
	statementChain.statements = make(map[string]*heartbeat.OpinionStatement)
	statementChain.tail = nil

	statementChain.Events.Reset.Trigger()
}

func (statementChain *StatementChain) GetTail() *heartbeat.OpinionStatement {
	return statementChain.tail
}
