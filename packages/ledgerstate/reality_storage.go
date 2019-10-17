package ledgerstate

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/errors"
)

type RealityStorage interface {
	LoadReality(realityId RealityId) (result *Reality, err errors.IdentifiableError)
	StoreReality(reality *Reality) (err errors.IdentifiableError)
}

type RealityStorageFactory func(id []byte) RealityStorage

type RealityStorageMemory struct {
	id        []byte
	realities map[RealityId]*Reality
	mutex     sync.RWMutex
}

var _ RealityStorage = &RealityStorageMemory{}

func newRealityStorageMemory(id []byte) RealityStorage {
	return &RealityStorageMemory{
		id:        id,
		realities: make(map[RealityId]*Reality),
	}
}

func (storage *RealityStorageMemory) StoreReality(reality *Reality) (err errors.IdentifiableError) {
	storage.mutex.Lock()

	storage.realities[reality.GetId()] = reality

	storage.mutex.Unlock()

	return
}

func (storage *RealityStorageMemory) LoadReality(realityId RealityId) (result *Reality, err errors.IdentifiableError) {
	storage.mutex.RLock()

	result = storage.realities[realityId]

	storage.mutex.RUnlock()

	return
}
