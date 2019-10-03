package ca

import (
	"bytes"
	"sort"
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
			ChainReset:       events.NewEvent(events.CallbackCaller),
			StatementMissing: events.NewEvent(HashCaller),
		},
		options:        DEFAULT_NEIGHBOR_MANAGER_OPTIONS.Override(options...),
		mainChain:      NewStatementChain(),
		neighborChains: make(map[string]*StatementChain),
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
		// cycle through heartbeats and apply them one by one
	}

	return
}

func (neighborManager *NeighborManager) ApplyHeartbeat(heartbeat *heartbeat.Heartbeat) (err errors.IdentifiableError) {
	if storeErr := neighborManager.storeHeartbeat(heartbeat); storeErr != nil {
		err = storeErr

		return
	}

	// region mark idle neighbors //////////////////////////////////////////////////////////////////////////////////////

	existingNeighbors := make(map[string]*StatementChain)
	for neighborId, neighborChain := range neighborManager.neighborChains {
		existingNeighbors[neighborId] = neighborChain
	}

	for neighborId := range neighborStatements {
		if neighborStatementChain, neighborExists := existingNeighbors[neighborId]; neighborExists {
			neighborStatementChain.ResetIdleCounter()

			delete(existingNeighbors, neighborId)
		}
	}

	for _, neighborStatementChain := range existingNeighbors {
		neighborStatementChain.IncreaseIdleCounter()
	}

	// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////

	// region add received statements //////////////////////////////////////////////////////////////////////////////////

	applyStatements := neighborManager.mainChain.AddStatement(mainStatement)

	// add neighbor statements to chain
	for neighborId, statementsOfNeighbor := range neighborStatements {
		neighborChain, exists := neighborManager.neighborChains[neighborId]
		if !exists {
			neighborChain = neighborManager.addNeighborChain(neighborId)
		}

		if neighborChain != nil {
			for _, neighborStatement := range statementsOfNeighbor {
				neighborChain.AddStatement(neighborStatement)
			}
		}
	}

	// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////

	// region apply pending statements /////////////////////////////////////////////////////////////////////////////////

	if applyStatements {
		neighborManager.mainChain.lastAppliedStatement = mainStatement
	}

	// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////

	return
}

func (neighborManager *NeighborManager) GenerateHeartbeatStatements() (result []*heartbeat.OpinionStatement) {
	result = make([]*heartbeat.OpinionStatement, 0)

	if lastAppliedStatement := neighborManager.mainChain.GetLastAppliedStatement(); lastAppliedStatement != nil {
		currentStatement := lastAppliedStatement
		for currentStatement != nil && !bytes.Equal(currentStatement.GetHash(), neighborManager.previouslyReportedHeartbeatHash) {
			result = append([]*heartbeat.OpinionStatement{currentStatement}, result...)

			currentStatement = neighborManager.mainChain.GetStatement(currentStatement.GetPreviousStatementHash())
		}

		neighborManager.previouslyReportedHeartbeatHash = lastAppliedStatement.GetHash()
	}

	return
}

func (neighborManager *NeighborManager) addNeighborChain(neighborId string) *StatementChain {
	if len(neighborManager.neighborChains) < MAX_NEIGHBOR_COUNT {
		newNeighborChain := NewStatementChain()
		neighborManager.neighborChains[neighborId] = newNeighborChain

		return newNeighborChain
	}

	neighbors := make([]string, 0, len(neighborManager.neighborChains))
	for neighborId := range neighborManager.neighborChains {
		neighbors = append(neighbors, neighborId)
	}

	sort.Slice(neighbors, func(i, j int) bool {
		neighborChainI := neighborManager.neighborChains[neighbors[i]]
		neighborChainJ := neighborManager.neighborChains[neighbors[i]]

		switch true {
		case neighborChainI.GetIdleCounter() < neighborChainJ.GetIdleCounter():
			return true
		default:
			return false

		}
	})
	newNeighborChain := NewStatementChain()

	neighborManager.neighborChains[neighborId] = newNeighborChain

	return newNeighborChain
}

type NeighborManagerOptions struct {
	maxNeighborChains int
}

func (options NeighborManagerOptions) Override(optionalOptions ...NeighborManagerOption) *NeighborManagerOptions {
	result := &options
	for _, option := range optionalOptions {
		option(result)
	}

	return result
}

type NeighborManagerOption func(*NeighborManagerOptions)
