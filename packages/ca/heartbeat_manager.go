package ca

import (
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/typeutils"

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
	droppedNeighbors [][]byte
	neighborManagers map[string]*NeighborManager
	opinions         *OpinionRegister

	droppedNeighborsMutex sync.RWMutex
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
		droppedNeighbors: make([][]byte, 0),
		neighborManagers: make(map[string]*NeighborManager),
		opinions:         NewOpinionRegister(),
	}
}

func (heartbeatManager *HeartbeatManager) AddNeighbor(neighborIdentity *identity.Identity) {
	heartbeatManager.neighborManagersMutex.RLock()
	if _, exists := heartbeatManager.neighborManagers[neighborIdentity.StringIdentifier]; !exists {
		heartbeatManager.neighborManagersMutex.RUnlock()

		heartbeatManager.neighborManagersMutex.Lock()
		if _, exists := heartbeatManager.neighborManagers[neighborIdentity.StringIdentifier]; !exists {
			newNeighborManager := NewNeighborManager()

			newNeighborManager.Events.UpdateOpinion.Attach(events.NewClosure(func(transactionId string, liked bool) {
				heartbeatManager.processNeighborOpinionUpdate(neighborIdentity, transactionId, liked)
			}))

			heartbeatManager.neighborManagers[neighborIdentity.StringIdentifier] = newNeighborManager
			heartbeatManager.neighborManagersMutex.Unlock()

			heartbeatManager.Events.AddNeighbor.Trigger(neighborIdentity, newNeighborManager)
		} else {
			heartbeatManager.neighborManagersMutex.Unlock()
		}
	} else {
		heartbeatManager.neighborManagersMutex.RUnlock()
	}
}

func (heartbeatManager *HeartbeatManager) RemoveNeighbor(neighborIdentity *identity.Identity) {
	heartbeatManager.neighborManagersMutex.RLock()
	if _, exists := heartbeatManager.neighborManagers[neighborIdentity.StringIdentifier]; exists {
		heartbeatManager.neighborManagersMutex.RUnlock()

		heartbeatManager.neighborManagersMutex.Lock()
		if neighborManager, exists := heartbeatManager.neighborManagers[neighborIdentity.StringIdentifier]; exists {
			delete(heartbeatManager.neighborManagers, neighborIdentity.StringIdentifier)
			heartbeatManager.neighborManagersMutex.Unlock()

			heartbeatManager.droppedNeighborsMutex.Lock()
			heartbeatManager.droppedNeighbors = append(heartbeatManager.droppedNeighbors, neighborIdentity.Identifier)
			heartbeatManager.droppedNeighborsMutex.Unlock()

			heartbeatManager.Events.RemoveNeighbor.Trigger(neighborIdentity, neighborManager)
		} else {
			heartbeatManager.neighborManagersMutex.Unlock()
		}
	} else {
		heartbeatManager.neighborManagersMutex.RUnlock()
	}
}

func (heartbeatManager *HeartbeatManager) InitialDislike(transactionId []byte) {
	heartbeatManager.opinions.CreateOpinion(typeutils.BytesToString(transactionId)).SetLiked(false)
}

func (heartbeatManager *HeartbeatManager) InitialLike(transactionId []byte) {
	heartbeatManager.opinions.CreateOpinion(typeutils.BytesToString(transactionId)).SetLiked(true)
}

func (heartbeatManager *HeartbeatManager) processNeighborOpinionUpdate(neighbor *identity.Identity, transactionId string, liked bool) {
	opinion := heartbeatManager.opinions.GetOpinion(transactionId)
	if !opinion.Exists() || opinion.IsLiked() != liked {
		totalWeight := len(heartbeatManager.neighborManagers)
		threshold := float64(totalWeight) / 2

		likedWeight := 0
		dislikedWeight := 0
		for _, neighborManager := range heartbeatManager.neighborManagers {
			weightOfNeighbor := 1

			if neighborOpinionLiked, exists := neighborManager.opinions[transactionId]; exists {
				if neighborOpinionLiked {
					likedWeight += weightOfNeighbor
				} else {
					dislikedWeight += weightOfNeighbor
				}
			}
		}

		if likedWeight > dislikedWeight && likedWeight > int(threshold) {
			if !opinion.Exists() || !opinion.IsLiked() {
				opinion = heartbeatManager.opinions.CreateOpinion(transactionId)
				opinion.SetLiked(true)
			}
		} else if dislikedWeight >= likedWeight && dislikedWeight >= int(threshold) {
			if !opinion.Exists() || opinion.IsLiked() {
				opinion = heartbeatManager.opinions.CreateOpinion(transactionId)
				opinion.SetLiked(false)
			}
		}
	}
}

func (heartbeatManager *HeartbeatManager) GenerateHeartbeat() (result *heartbeat.Heartbeat, err errors.IdentifiableError) {
	mainStatement, mainStatementErr := heartbeatManager.generateMainStatement()
	if mainStatementErr != nil {
		err = mainStatementErr

		return
	}

	neighborStatements, neighborStatementErr := heartbeatManager.generateNeighborStatements()
	if neighborStatementErr != nil {
		err = neighborStatementErr

		return
	}

	generatedHeartbeat := heartbeat.NewHeartbeat()
	generatedHeartbeat.SetNodeId(heartbeatManager.identity.StringIdentifier)
	generatedHeartbeat.SetMainStatement(mainStatement)
	generatedHeartbeat.SetDroppedNeighbors(heartbeatManager.generateDroppedNeighbors())
	generatedHeartbeat.SetNeighborStatements(neighborStatements)

	signingErr := generatedHeartbeat.Sign(heartbeatManager.identity)
	if signingErr != nil {
		err = signingErr
	}

	result = generatedHeartbeat

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

	if lastAppliedStatement := heartbeatManager.statementChain.GetTail(); lastAppliedStatement != nil {
		mainStatement.SetPreviousStatementHash(lastAppliedStatement.GetHash())
	}

	if signingErr := mainStatement.Sign(heartbeatManager.identity); signingErr == nil {
		result = mainStatement

		heartbeatManager.opinions.ApplyPendingOpinions()
		heartbeatManager.statementChain.tail = mainStatement
	} else {
		err = signingErr
	}

	return
}

func (heartbeatManager *HeartbeatManager) generateDroppedNeighbors() (result [][]byte) {
	heartbeatManager.droppedNeighborsMutex.RLock()
	result = make([][]byte, len(heartbeatManager.droppedNeighbors))
	copy(result, heartbeatManager.droppedNeighbors)
	heartbeatManager.droppedNeighborsMutex.RUnlock()

	heartbeatManager.droppedNeighborsMutex.Lock()
	heartbeatManager.droppedNeighbors = make([][]byte, 0)
	heartbeatManager.droppedNeighborsMutex.Unlock()

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
	for transactionId, opinion := range heartbeatManager.opinions.GetPendingOpinions() {
		if !opinion.IsLiked() {
			newToggledTransaction := heartbeat.NewToggledTransaction()
			newToggledTransaction.SetInitialStatement(opinion.IsInitial())
			newToggledTransaction.SetFinalStatement(opinion.IsFinalized())
			newToggledTransaction.SetTransactionId([]byte(transactionId))

			toggledTransactions = append(toggledTransactions, newToggledTransaction)
		}
	}

	return toggledTransactions
}
