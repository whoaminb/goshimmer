package ca

import (
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/events"

	"github.com/iotaledger/goshimmer/packages/identity"

	"github.com/iotaledger/goshimmer/packages/ca/heartbeat"

	"github.com/iotaledger/goshimmer/packages/errors"
)

type HeartbeatManager struct {
	Events *HeartbeatManagerEvents

	identity         *identity.Identity
	options          *HeartbeatManagerOptions
	statementChain   *StatementChain
	neighborManagers map[string]*NeighborManager
	initialOpinions  map[string]bool

	neighborManagersMutex sync.RWMutex
}

func NewHeartbeatManager(identity *identity.Identity, options ...HeartbeatManagerOption) *HeartbeatManager {
	return &HeartbeatManager{
		Events: &HeartbeatManagerEvents{
			AddNeighbor:    events.NewEvent(IdentityNeighborManagerCaller),
			RemoveNeighbor: events.NewEvent(IdentityNeighborManagerCaller),
		},

		identity: identity,
		options:  DEFAULT_OPTIONS.Override(options...),

		statementChain:   NewStatementChain(),
		neighborManagers: make(map[string]*NeighborManager),
		initialOpinions:  make(map[string]bool),
	}
}

func (heartbeatManager *HeartbeatManager) AddNeighbor(neighborIdentity *identity.Identity) {
	if _, exists := heartbeatManager.neighborManagers[neighborIdentity.StringIdentifier]; !exists {
		newNeighborManager := NewNeighborManager()

		heartbeatManager.neighborManagers[neighborIdentity.StringIdentifier] = newNeighborManager

		heartbeatManager.Events.AddNeighbor.Trigger(neighborIdentity, newNeighborManager)
	}
}

func (heartbeatManager *HeartbeatManager) InitialDislike(transactionId []byte) {
	heartbeatManager.initialOpinions[string(transactionId)] = false
}

func (heartbeatManager *HeartbeatManager) InitialLike(transactionId []byte) {
	heartbeatManager.initialOpinions[string(transactionId)] = true
}

func (heartbeatManager *HeartbeatManager) GenerateHeartbeat() (result *heartbeat.Heartbeat, err errors.IdentifiableError) {
	if mainStatement, mainStatementErr := heartbeatManager.generateMainStatement(); mainStatementErr == nil {
		if neighborStatements, neighborStatementErr := heartbeatManager.generateNeighborStatements(); neighborStatementErr == nil {
			generatedHeartbeat := heartbeat.NewHeartbeat()
			generatedHeartbeat.SetNodeId(heartbeatManager.identity.StringIdentifier)
			generatedHeartbeat.SetMainStatement(mainStatement)
			generatedHeartbeat.SetNeighborStatements(neighborStatements)

			if signingErr := generatedHeartbeat.Sign(heartbeatManager.identity); signingErr == nil {
				result = generatedHeartbeat
			} else {
				err = signingErr
			}
		} else {
			err = neighborStatementErr
		}
	} else {
		err = mainStatementErr
	}

	return
}

func (heartbeatManager *HeartbeatManager) ApplyHeartbeat(heartbeat *heartbeat.Heartbeat) (err errors.IdentifiableError) {
	heartbeatManager.neighborManagersMutex.RLock()
	defer heartbeatManager.neighborManagersMutex.RUnlock()

	if signatureValid, signatureErr := heartbeat.VerifySignature(); signatureErr == nil {
		if signatureValid {
			issuerId := heartbeat.GetNodeId()

			neighborManager, neighborExists := heartbeatManager.neighborManagers[issuerId]
			if !neighborExists {
				err = ErrUnknownNeighbor.Derive("unknown neighbor: " + issuerId)
			} else {
				err = neighborManager.ApplyHeartbeat(heartbeat)
			}
		} else {
			err = ErrMalformedHeartbeat.Derive("the heartbeat has an invalid signature")
		}
	} else {
		err = signatureErr
	}

	return
}

func (heartbeatManager *HeartbeatManager) generateMainStatement() (result *heartbeat.OpinionStatement, err errors.IdentifiableError) {
	mainStatement := heartbeat.NewOpinionStatement()
	mainStatement.SetNodeId(heartbeatManager.identity.StringIdentifier)
	mainStatement.SetTime(uint64(time.Now().Unix()))
	mainStatement.SetToggledTransactions(heartbeatManager.generateToggledTransactions())

	if lastAppliedStatement := heartbeatManager.statementChain.lastAppliedStatement; lastAppliedStatement != nil {
		mainStatement.SetPreviousStatementHash(lastAppliedStatement.GetHash())
	}

	if signingErr := mainStatement.Sign(heartbeatManager.identity); signingErr == nil {
		result = mainStatement

		heartbeatManager.resetInitialOpinions()
		heartbeatManager.statementChain.lastAppliedStatement = mainStatement
	} else {
		err = signingErr
	}

	return
}

func (heartbeatManager *HeartbeatManager) generateNeighborStatements() (result map[string][]*heartbeat.OpinionStatement, err errors.IdentifiableError) {
	result = make(map[string][]*heartbeat.OpinionStatement)

	for neighborId, neighborManager := range heartbeatManager.neighborManagers {
		if lastAppliedStatements := neighborManager.GenerateHeartbeatStatements(); len(lastAppliedStatements) >= 1 {
			result[neighborId] = lastAppliedStatements
		}
	}

	return
}

func (heartbeatManager *HeartbeatManager) generateToggledTransactions() []*heartbeat.ToggledTransaction {
	toggledTransactions := make([]*heartbeat.ToggledTransaction, 0)
	for transactionId, liked := range heartbeatManager.initialOpinions {
		if !liked {
			newToggledTransaction := heartbeat.NewToggledTransaction()
			newToggledTransaction.SetInitialStatement(true)
			newToggledTransaction.SetFinalStatement(false)
			newToggledTransaction.SetTransactionId([]byte(transactionId))

			toggledTransactions = append(toggledTransactions, newToggledTransaction)
		}
	}

	return toggledTransactions
}

func (heartbeatManager *HeartbeatManager) resetInitialOpinions() {
	heartbeatManager.initialOpinions = make(map[string]bool)
}
