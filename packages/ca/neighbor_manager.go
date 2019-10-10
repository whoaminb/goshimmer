package ca

import (
	"bytes"
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
	opinions                        map[string]bool
	previouslyReportedHeartbeatHash []byte
}

func NewNeighborManager(options ...NeighborManagerOption) *NeighborManager {
	return &NeighborManager{
		Events: NeighborManagerEvents{
			AddNeighbor: events.NewEvent(func(handler interface{}, params ...interface{}) {

			}),
			RemoveNeighbor: events.NewEvent(func(handler interface{}, params ...interface{}) {

			}),
			NeighborActive: events.NewEvent(func(handler interface{}, params ...interface{}) {

			}),
			NeighborIdle: events.NewEvent(func(handler interface{}, params ...interface{}) {

			}),
			ChainReset:       events.NewEvent(events.CallbackCaller),
			StatementMissing: events.NewEvent(HashCaller),
			UpdateOpinion:    events.NewEvent(StringBoolCaller),
		},
		options:           DEFAULT_NEIGHBOR_MANAGER_OPTIONS.Override(options...),
		mainChain:         NewStatementChain(),
		missingHeartbeats: make(map[string]bool),
		pendingHeartbeats: make(map[string]*heartbeat.Heartbeat),
		heartbeats:        make(map[string]*heartbeat.Heartbeat),
		opinions:          make(map[string]bool),
		neighborChains:    make(map[string]*StatementChain),
	}
}

func (neighborManager *NeighborManager) Reset() {
	neighborManager.mainChain.Reset()
	neighborManager.neighborChains = make(map[string]*StatementChain)
	neighborManager.opinions = make(map[string]bool)
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
			if applicationErr := neighborManager.applySolidHeartbeat(sortedHeartbeat); applicationErr != nil {
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
			neighborManager.Events.NeighborActive.Trigger(neighborId)

			delete(idleNeighbors, neighborId)
		}
	}

	for neighborId := range idleNeighbors {
		neighborManager.Events.NeighborIdle.Trigger(neighborId)
	}
}

func (neighborManager *NeighborManager) updateStatementChains(mainStatement *heartbeat.OpinionStatement, neighborStatements map[string][]*heartbeat.OpinionStatement) (err errors.IdentifiableError) {
	if err = neighborManager.mainChain.AddStatement(mainStatement); err != nil {
		return
	}

	for neighborId, statementsOfNeighbor := range neighborStatements {
		neighborChain, neighborChainErr := neighborManager.addNeighborChain(neighborId)
		if neighborChainErr != nil {
			err = neighborChainErr

			return
		}

		for _, neighborStatement := range statementsOfNeighbor {
			if statementErr := neighborChain.AddStatement(neighborStatement); statementErr != nil {
				err = statementErr

				return
			}
		}
	}

	return
}

func (neighborManager *NeighborManager) applySolidHeartbeat(heartbeat *heartbeat.Heartbeat) (err errors.IdentifiableError) {
	mainStatement := heartbeat.GetMainStatement()
	neighborStatements := heartbeat.GetNeighborStatements()

	neighborManager.removeDroppedNeighbors(heartbeat.GetDroppedNeighbors())
	neighborManager.markIdleNeighbors(neighborStatements)
	if err = neighborManager.updateStatementChains(mainStatement, neighborStatements); err != nil {
		return
	}
	if err = neighborManager.updateNeighborManager(); err != nil {
		return
	}

	return
}

func (neighborManager *NeighborManager) updateNeighborManager() (err errors.IdentifiableError) {
	updatedOpinions, verificationErr := neighborManager.retrieveAndVerifyUpdates()
	if verificationErr != nil {
		err = verificationErr

		return
	}

	for transactionId, liked := range updatedOpinions {
		if currentlyLiked, opinionExists := neighborManager.opinions[transactionId]; !opinionExists || currentlyLiked != liked {
			neighborManager.opinions[transactionId] = liked

			neighborManager.Events.UpdateOpinion.Trigger(transactionId, liked)
		}
	}

	return
}

func (neighborManager *NeighborManager) retrieveAndVerifyUpdates() (updates map[string]bool, err errors.IdentifiableError) {
	updates = make(map[string]bool)

	// retrieve required parameters
	totalWeight := len(neighborManager.neighborChains)
	threshold := float64(totalWeight) / 2
	opinionsOfNeighbors := neighborManager.getAccumulatedPendingOpinionsOfNeighbors()

	mainChainInitialOpinions := make(map[string]*Opinion)
	mainChainOpinionChanges := make(map[string]*Opinion)
	for transactionId, opinion := range neighborManager.mainChain.GetOpinions().GetPendingOpinions() {
		if opinion.IsInitial() {
			mainChainInitialOpinions[transactionId] = opinion
		} else {
			mainChainOpinionChanges[transactionId] = opinion
		}
	}
	// always consider initial opinions
	for transactionId, opinion := range mainChainInitialOpinions {
		updates[transactionId] = opinion.IsLiked()
	}

	// consider opinions that have seen enough neighbors
	for transactionId, opinion := range opinionsOfNeighbors {
		if opinion[0] > opinion[1] && opinion[0] > int(threshold) {
			// main statement "should" like it
			if changedOpinion := mainChainOpinionChanges[transactionId]; !changedOpinion.Exists() || !changedOpinion.IsLiked() {
				if initialOpinion := mainChainInitialOpinions[transactionId]; !initialOpinion.Exists() || !initialOpinion.IsLiked() {
					err = ErrMalformedHeartbeat.Derive("main statement should like transaction")

					return
				} else {
					updates[transactionId] = true
				}
			} else {
				updates[transactionId] = true
			}
		} else if opinion[1] > opinion[0] && opinion[1] >= int(threshold) {
			// main statement "should" dislike it
			if changedOpinion := mainChainOpinionChanges[transactionId]; !changedOpinion.Exists() || changedOpinion.IsLiked() {
				if initialOpinion := mainChainInitialOpinions[transactionId]; !initialOpinion.Exists() || initialOpinion.IsLiked() {
					err = ErrMalformedHeartbeat.Derive("main statement should dislike transaction")

					return
				} else {
					updates[transactionId] = false
				}
			} else {
				updates[transactionId] = false
			}
		}
	}

	return
}

func (neighborManager *NeighborManager) getAccumulatedPendingOpinionsOfNeighbors() (result map[string][]int) {
	result = make(map[string][]int)

	for _, neighborChain := range neighborManager.neighborChains {
		for transactionId, pendingOpinion := range neighborChain.GetOpinions().GetPendingOpinions() {
			opinion, exists := result[transactionId]
			if !exists {
				opinion = make([]int, 2)

				result[transactionId] = opinion
			}

			weightOfNeighbor := 1
			if pendingOpinion.IsLiked() {
				opinion[0] += weightOfNeighbor
			} else {
				opinion[1] += weightOfNeighbor
			}
		}
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
