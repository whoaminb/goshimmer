package ca

import (
	"sync"
)

type Opinion struct {
	initial   bool
	liked     bool
	finalized bool
	pending   bool

	initialMutex   sync.RWMutex
	likedMutex     sync.RWMutex
	finalizedMutex sync.RWMutex
	pendingMutex   sync.RWMutex
}

func NewOpinion() *Opinion {
	return &Opinion{}
}

func (opinion *Opinion) IsLiked() bool {
	opinion.likedMutex.RLock()
	defer opinion.likedMutex.RUnlock()

	return opinion.liked
}

func (opinion *Opinion) SetLiked(liked bool) *Opinion {
	opinion.likedMutex.Lock()
	defer opinion.likedMutex.Unlock()

	opinion.liked = liked

	return opinion
}

func (opinion *Opinion) IsInitial() bool {
	opinion.initialMutex.RLock()
	defer opinion.initialMutex.RLock()

	return opinion.initial
}

func (opinion *Opinion) SetInitial(initial bool) *Opinion {
	opinion.initialMutex.Lock()
	defer opinion.initialMutex.Unlock()

	opinion.initial = initial

	return opinion
}

func (opinion *Opinion) IsFinalized() bool {
	opinion.finalizedMutex.RLock()
	defer opinion.finalizedMutex.RLock()

	return opinion.finalized
}

func (opinion *Opinion) SetFinalized(finalized bool) *Opinion {
	opinion.finalizedMutex.Lock()
	defer opinion.finalizedMutex.Unlock()

	opinion.finalized = finalized

	return opinion
}

func (opinion *Opinion) IsPending() bool {
	opinion.pendingMutex.RLock()
	defer opinion.pendingMutex.RUnlock()

	return opinion.pending
}

func (opinion *Opinion) SetPending(pending bool) *Opinion {
	opinion.pendingMutex.Lock()
	defer opinion.pendingMutex.Unlock()

	opinion.pending = pending

	return opinion
}

func (opinion *Opinion) Exists() bool {
	return opinion != nil
}
