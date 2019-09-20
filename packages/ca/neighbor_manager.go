package ca

import (
	"bytes"
	"strconv"

	"github.com/iotaledger/goshimmer/packages/ca/heartbeat"

	"github.com/iotaledger/goshimmer/packages/errors"
)

type NeighborManager struct {
	options        *NeighborManagerOptions
	mainChain      *StatementChain
	neighborChains map[string]*StatementChain
}

func NewNeighborManager(options ...NeighborManagerOption) *NeighborManager {
	return &NeighborManager{
		options:        (&NeighborManagerOptions{}).Override(options...),
		mainChain:      NewStatementChain(),
		neighborChains: make(map[string]*StatementChain),
	}
}

func (neighborManager *NeighborManager) Reset() {
	neighborManager.mainChain.Reset()
	neighborManager.neighborChains = make(map[string]*StatementChain)
}

func (neighborManager *NeighborManager) ApplyHeartbeat(heartbeat *heartbeat.Heartbeat) (err errors.IdentifiableError) {
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

	// region check if heartbeat is semantically correct ///////////////////////////////////////////////////////////////

	// check if referenced main statement is missing
	if previousStatementHash := mainStatement.GetPreviousStatementHash(); len(previousStatementHash) != 0 {
		lastAppliedMainStatement := neighborManager.mainChain.GetLastAppliedStatement()
		if lastAppliedMainStatement != nil && !bytes.Equal(lastAppliedMainStatement.GetHash(), previousStatementHash) {
			if (mainStatement.GetTime() - lastAppliedMainStatement.GetTime()) < MAX_STATEMENT_TIMEOUT {
				// request missing heartbeat
				neighborManager.mainChain.AddStatement(mainStatement)

				return
			} else {
				neighborManager.Reset()
			}
		}
	} else {
		neighborManager.mainChain.AddStatement(mainStatement)
	}

	// check if referenced neighbor statements are missing
	for neighborId, statementsOfNeighbor := range neighborStatements {
		neighborChain, exists := neighborManager.neighborChains[neighborId]
		if exists {
			for _, neighborStatement := range statementsOfNeighbor {
				lastAppliedNeighborStatement := neighborChain.GetLastAppliedStatement()
				if lastAppliedNeighborStatement != nil && !bytes.Equal(lastAppliedNeighborStatement.GetHash(), neighborStatement.GetPreviousStatementHash()) {
					return ErrMalformedHeartbeat.Derive("missing neighbor statement")
				}
			}
		} else {
			// 1. check if new slot is available (not full || statement of neighbor with last connection)
		}
	}

	// check if neighbor statements are a valid justification for main statement
	// sthsth

	// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////

	return
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
