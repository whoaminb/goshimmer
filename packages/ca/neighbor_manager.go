package ca

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/iotaledger/goshimmer/packages/typeutils"

	"github.com/iotaledger/goshimmer/packages/events"

	"github.com/iotaledger/goshimmer/packages/ca/heartbeat"

	"github.com/iotaledger/goshimmer/packages/errors"
)

type NeighborManager struct {
	Events                          NeighborManagerEvents
	options                         *NeighborManagerOptions
	lastReceivedHeartbeat           *heartbeat.Heartbeat
	missingHeartbeats               map[string]bool
	pendingHeartbeats               map[string]*heartbeat.Heartbeat
	heartbeats                      map[string]*heartbeat.Heartbeat
	mainChain                       *StatementChain
	neighborChains                  map[string]*StatementChain
	previouslyReportedHeartbeatHash []byte
}

func NewNeighborManager(options ...NeighborManagerOption) *NeighborManager {
	return &NeighborManager{
		Events: NeighborManagerEvents{
			AddNeighbor: events.NewEvent(func(handler interface{}, params ...interface{}) {

			}),
			RemoveNeighbor: events.NewEvent(func(handler interface{}, params ...interface{}) {

			}),
			ChainReset:       events.NewEvent(events.CallbackCaller),
			StatementMissing: events.NewEvent(HashCaller),
		},
		options:           DEFAULT_NEIGHBOR_MANAGER_OPTIONS.Override(options...),
		mainChain:         NewStatementChain(),
		missingHeartbeats: make(map[string]bool),
		pendingHeartbeats: make(map[string]*heartbeat.Heartbeat),
		heartbeats:        make(map[string]*heartbeat.Heartbeat),
		neighborChains:    make(map[string]*StatementChain),
	}
}

func (neighborManager *NeighborManager) Reset() {
	neighborManager.mainChain.Reset()
	neighborManager.neighborChains = make(map[string]*StatementChain)
}

func (neighborManager *NeighborManager) storeHeartbeat(heartbeat *heartbeat.Heartbeat) (err errors.IdentifiableError) {
	// region check if heartbeat is "syntactically correct" ////////////////////////////////////////////////////////////

	mainStatement := heartbeat.GetMainStatement()
	if mainStatement == nil {
		return ErrMalformedHeartbeat.Derive("missing main statement in heartbeat")
	}

	neighborStatements := heartbeat.GetNeighborStatements()
	if len(neighborStatements) > neighborManager.options.maxNeighborChains {
		return ErrTooManyNeighbors.Derive("too many neighbors in statement of " + heartbeat.GetNodeId() + ": " + strconv.Itoa(len(neighborStatements)))
	}

	// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////

	// region check if heartbeat is solid //////////////////////////////////////////////////////////////////////////////

	previousHeartbeatHash := mainStatement.GetPreviousStatementHash()
	if len(previousHeartbeatHash) == 0 {
		neighborManager.Reset()
	} else if neighborManager.lastReceivedHeartbeat != nil {
		lastMainStatement := neighborManager.lastReceivedHeartbeat.GetMainStatement()
		previousHeartbeatHashString := typeutils.BytesToString(previousHeartbeatHash)
		if lastMainStatement != nil && !bytes.Equal(lastMainStatement.GetHash(), previousHeartbeatHash) {
			if len(neighborManager.pendingHeartbeats) >= MAX_PENDING_HEARTBEATS || len(neighborManager.missingHeartbeats) >= MAX_MISSING_HEARTBEATS {
				neighborManager.Reset()
			} else if _, exists := neighborManager.heartbeats[previousHeartbeatHashString]; !exists {
				neighborManager.missingHeartbeats[previousHeartbeatHashString] = true
			}
		}
	}

	// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////

	// region store heartbeat and update lists /////////////////////////////////////////////////////////////////////////

	heartbeatHash := typeutils.BytesToString(mainStatement.GetHash())

	if neighborManager.lastReceivedHeartbeat == nil || mainStatement.GetTime() > neighborManager.lastReceivedHeartbeat.GetMainStatement().GetTime() {
		neighborManager.lastReceivedHeartbeat = heartbeat
	}

	neighborManager.heartbeats[heartbeatHash] = heartbeat
	neighborManager.pendingHeartbeats[heartbeatHash] = heartbeat
	delete(neighborManager.missingHeartbeats, heartbeatHash)

	// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////

	return
}

func (neighborManager *NeighborManager) applyPendingHeartbeats() (err errors.IdentifiableError) {
	if len(neighborManager.missingHeartbeats) == 0 && len(neighborManager.pendingHeartbeats) >= 1 {
		sortedPendingHeartbeats, sortingErr := neighborManager.getPendingHeartbeatsSorted()
		if sortingErr != nil {
			err = sortingErr

			return
		}

		for _, sortedHeartbeat := range sortedPendingHeartbeats {
			if applicationErr := neighborManager.applyHeartbeat(sortedHeartbeat); applicationErr != nil {
				err = applicationErr

				return
			}
		}
	}

	return
}

func (neighborManager *NeighborManager) getPendingHeartbeatsSorted() (result []*heartbeat.Heartbeat, err errors.IdentifiableError) {
	pendingHeartbeatCount := len(neighborManager.pendingHeartbeats)
	result = make([]*heartbeat.Heartbeat, pendingHeartbeatCount)

	processedHeartbeats := 0
	currentHeartbeat := neighborManager.lastReceivedHeartbeat
	for currentHeartbeat != nil && len(neighborManager.pendingHeartbeats) >= 1 {
		mainStatement := currentHeartbeat.GetMainStatement()
		if mainStatement == nil {
			result = nil
			err = ErrInternalError.Derive("missing main statement in heartbeat")

			return
		}

		currentHeartbeatHash := typeutils.BytesToString(mainStatement.GetHash())
		previousHeartbeatHash := typeutils.BytesToString(mainStatement.GetPreviousStatementHash())

		if _, exists := neighborManager.pendingHeartbeats[currentHeartbeatHash]; !exists {
			result = nil
			err = ErrInternalError.Derive("pending heartbeats list is out of sync")

			return
		}
		delete(neighborManager.pendingHeartbeats, currentHeartbeatHash)

		result[pendingHeartbeatCount-processedHeartbeats-1] = currentHeartbeat

		currentHeartbeat = neighborManager.heartbeats[previousHeartbeatHash]
		processedHeartbeats++
	}

	return
}

// this method removes dropped neighbors from the neighborChains
func (neighborManager *NeighborManager) removeDroppedNeighbors(droppedNeighbors [][]byte) {
	for _, droppedNeighbor := range droppedNeighbors {
		neighborIdString := typeutils.BytesToString(droppedNeighbor)

		if _, exists := neighborManager.neighborChains[neighborIdString]; exists {
			delete(neighborManager.neighborChains, neighborIdString)

			neighborManager.Events.RemoveNeighbor.Trigger(droppedNeighbor)
		}
	}
}

func (neighborManager *NeighborManager) markIdleNeighbors(neighborStatements map[string][]*heartbeat.OpinionStatement) {
	idleNeighbors := make(map[string]*StatementChain)
	for neighborId, neighborChain := range neighborManager.neighborChains {
		idleNeighbors[neighborId] = neighborChain
	}

	for neighborId := range neighborStatements {
		if _, neighborExists := idleNeighbors[neighborId]; neighborExists {
			// TRIGGER ACTIVE

			delete(idleNeighbors, neighborId)
		}
	}

	for _, x := range idleNeighbors {
		// TRIGGER IDLE
		if false {
			fmt.Println(x)
		}
	}
}

func (neighborManager *NeighborManager) updateStatementChains(mainStatement *heartbeat.OpinionStatement, neighborStatements map[string][]*heartbeat.OpinionStatement) (err errors.IdentifiableError) {
	neighborManager.mainChain.AddStatement(mainStatement)

	for neighborId, statementsOfNeighbor := range neighborStatements {
		neighborChain, neighborChainErr := neighborManager.addNeighborChain(neighborId)
		if neighborChainErr != nil {
			err = neighborChainErr

			return
		}

		for _, neighborStatement := range statementsOfNeighbor {
			neighborChain.AddStatement(neighborStatement)
		}
	}

	return
}

func (neighborManager *NeighborManager) applyHeartbeat(heartbeat *heartbeat.Heartbeat) (err errors.IdentifiableError) {
	mainStatement := heartbeat.GetMainStatement()
	neighborStatements := heartbeat.GetNeighborStatements()

	neighborManager.removeDroppedNeighbors(heartbeat.GetDroppedNeighbors())

	neighborManager.markIdleNeighbors(neighborStatements)

	if err = neighborManager.updateStatementChains(mainStatement, neighborStatements); err != nil {
		return
	}

	return
}

func (neighborManager *NeighborManager) ApplyHeartbeat(heartbeat *heartbeat.Heartbeat) (err errors.IdentifiableError) {
	if storeErr := neighborManager.storeHeartbeat(heartbeat); storeErr != nil {
		err = storeErr

		return
	}

	if applicationErr := neighborManager.applyPendingHeartbeats(); applicationErr != nil {
		err = applicationErr

		return
	}

	return
}

func (neighborManager *NeighborManager) GenerateHeartbeatStatements() (result []*heartbeat.OpinionStatement) {
	result = make([]*heartbeat.OpinionStatement, 0)

	if lastAppliedStatement := neighborManager.mainChain.GetTail(); lastAppliedStatement != nil {
		currentStatement := lastAppliedStatement
		for currentStatement != nil && !bytes.Equal(currentStatement.GetHash(), neighborManager.previouslyReportedHeartbeatHash) {
			result = append([]*heartbeat.OpinionStatement{currentStatement}, result...)

			currentStatement = neighborManager.mainChain.GetStatement(currentStatement.GetPreviousStatementHash())
		}

		neighborManager.previouslyReportedHeartbeatHash = lastAppliedStatement.GetHash()
	}

	return
}

func (neighborManager *NeighborManager) addNeighborChain(neighborId string) (result *StatementChain, err errors.IdentifiableError) {
	if existingNeighborChain, exists := neighborManager.neighborChains[neighborId]; exists {
		result = existingNeighborChain

		return
	}

	if len(neighborManager.neighborChains) >= MAX_NEIGHBOR_COUNT {
		err = ErrTooManyNeighbors.Derive("failed to add new neighbor: too many neighbors")

		return
	}

	result = NewStatementChain()
	neighborManager.neighborChains[neighborId] = result

	neighborManager.Events.AddNeighbor.Trigger(neighborId, result)

	return
}
