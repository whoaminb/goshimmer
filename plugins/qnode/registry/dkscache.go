package registry

import (
	"github.com/iotaledger/goshimmer/packages/binary/valuetransfer/address"
	"github.com/iotaledger/goshimmer/plugins/qnode/tcrypto"
	"sync"
)

var (
	dkscache      = make(map[address.Address]*tcrypto.DKShare)
	dkscacheMutex = &sync.RWMutex{}
)

func CacheDKShare(dkshare *tcrypto.DKShare) {
	dkscacheMutex.Lock()
	defer dkscacheMutex.Unlock()
	dkscache[*dkshare.Address] = dkshare
}

func UncacheDKShare(addr *address.Address) {
	dkscacheMutex.Lock()
	defer dkscacheMutex.Unlock()
	delete(dkscache, *addr)
}

func GetDKShare(addr *address.Address) (*tcrypto.DKShare, bool, error) {
	dkscacheMutex.RLock()
	ret, ok := dkscache[*addr]
	if ok {
		defer dkscacheMutex.RUnlock()
		return ret, true, nil
	}
	// switching to write lock
	dkscacheMutex.RUnlock()
	dkscacheMutex.Lock()
	defer dkscacheMutex.Lock()

	var err error
	ok, err = ExistDKShareInRegistry(addr)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	ks, err := LoadDKShare(addr, false)
	if err != nil {
		return nil, false, err
	}
	dkscache[*addr] = ks
	return ks, true, nil
}
