package ram

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/interfaces"

	"github.com/iotaledger/goshimmer/packages/errors"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
)

type RealityStorage struct {
	id        []byte
	realities map[hash.Reality]interfaces.Reality
	mutex     sync.RWMutex
}

func NewRealityStorage(id []byte) interfaces.RealityStorage {
	return &RealityStorage{
		id:        id,
		realities: make(map[hash.Reality]interfaces.Reality),
	}
}

func (storage *RealityStorage) StoreReality(reality interfaces.Reality) (err errors.IdentifiableError) {
	storage.mutex.Lock()

	storage.realities[reality.GetId()] = reality

	storage.mutex.Unlock()

	return
}

func (storage *RealityStorage) LoadReality(realityId hash.Reality) (result interfaces.Reality, err errors.IdentifiableError) {
	storage.mutex.RLock()

	result = storage.realities[realityId]

	storage.mutex.RUnlock()

	return
}
