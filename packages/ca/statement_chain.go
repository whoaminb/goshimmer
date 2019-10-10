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

	opinions   *OpinionRegister
	statements map[string]*heartbeat.OpinionStatement
	tail       *heartbeat.OpinionStatement
}

func NewStatementChain() *StatementChain {
	return &StatementChain{
		Events: StatementChainEvents{
			Reset: events.NewEvent(events.CallbackCaller),
		},
		opinions:   NewOpinionRegister(),
		statements: make(map[string]*heartbeat.OpinionStatement),
	}
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
			opinion := statementChain.opinions.GetOpinion(transactionId)
			if opinion.Exists() {
				return ErrMalformedHeartbeat.Derive("two initial statements for the same transaction")
			}

			statementChain.opinions.CreateOpinion(transactionId).SetLiked(false)
		} else if toggledTransaction.IsFinalStatement() {
			// finalize -> clean up
		} else {
			opinion := statementChain.opinions.GetOpinion(transactionId)
			if opinion.Exists() {
				if opinion.IsPending() {
					return ErrMalformedHeartbeat.Derive("two changed statements for the same transaction")
				}

				opinion.SetInitial(false)
				opinion.SetPending(true)

				statementChain.opinions.pendingOpinions[transactionId] = opinion
			}
		}
	}

	statementChain.statements[typeutils.BytesToString(statement.GetHash())] = statement
	statementChain.tail = statement

	return nil
}

func (statementChain *StatementChain) ApplyPendingTransactionStatusChanges() {
	statementChain.opinions.ApplyPendingOpinions()
}

func (statementChain *StatementChain) GetStatement(statementHash []byte) *heartbeat.OpinionStatement {
	return statementChain.statements[typeutils.BytesToString(statementHash)]
}

func (statementChain *StatementChain) Reset() {
	statementChain.opinions = NewOpinionRegister()
	statementChain.statements = make(map[string]*heartbeat.OpinionStatement)
	statementChain.tail = nil

	statementChain.Events.Reset.Trigger()
}

func (statementChain *StatementChain) GetTail() *heartbeat.OpinionStatement {
	return statementChain.tail
}

func (statementChain *StatementChain) GetOpinions() *OpinionRegister {
	return statementChain.opinions
}
