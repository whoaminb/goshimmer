// Package defines interface to the persistent registry of the qnode
// The registry stores information about smart contracts and private keys and other data needed
// to sign the transaction
// all registry is cached in memory to enable fast check is SC transaction is of interest fo the node
// only SCData records which node is processing is included in the cache
// if scid is not in cache, the transaction is ignored
package registry

import (
	"bytes"
	"encoding/json"
	"github.com/iotaledger/goshimmer/plugins/qnode/db"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/transaction"
	"github.com/iotaledger/hive.go/database"
	"sync"
)

var (
	scDataCache map[transaction.ScId]*SCData
	scDataMutex = &sync.RWMutex{}
)

// SCData represents information on the SC and the committee, available to the node
// scid contains hash of the origin tx and the sc account address
type SCData struct {
	ScId          *transaction.ScId `json:"scid"`
	OwnerPubKey   *HashValue        `json:"owner_pub_key"`
	Description   string            `json:"description"`
	ProgramHash   *HashValue        `json:"program_hash"`
	NodeLocations []*PortAddr       `json:"node_locations"`
}

// RefreshSCDataCache reads all available SCDataRecords and puts them into the cache
// It filters out records which doesn't contain own address
func RefreshSCDataCache(ownAddr *PortAddr) error {
	scDataMutex.Lock()
	defer scDataMutex.Unlock()

	// clear or create the cache
	scDataCache = make(map[transaction.ScId]*SCData)
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	err = dbase.ForEachPrefix(dbSCDataGroupPrefix(), func(entry database.Entry) bool {
		scdata := &SCData{}
		if err = json.Unmarshal(entry.Value, scdata); err == nil {
			validateAndAddToCache(scdata, ownAddr)
		}
		return false
	})
	return err
}

// validates and adds SCData record to cache
func validateAndAddToCache(scdata *SCData, ownAddr *PortAddr) {
	addr := scdata.ScId.Address()
	dkshare, ok, _ := GetDKShare(&addr)
	if !ok {
		// failed to load dkshare of the sc address
		return
	}
	if int(dkshare.Index) >= len(scdata.NodeLocations) {
		// shouldn't be
		return
	}
	if ownAddr.String() != scdata.NodeLocations[dkshare.Index].String() {
		// if own address is not consistent with the one at the index in the list of nodes
		// this node is not interested in the SC
		return
	}
	scDataCache[*scdata.ScId] = scdata
}

// GetScData retrieves SCdata from the cache if it is there
func GetScData(scid *transaction.ScId) (*SCData, bool) {
	scDataMutex.RLock()
	defer scDataMutex.RUnlock()
	ret, ok := scDataCache[*scid]
	if !ok {
		return nil, false
	}
	return ret, true
}

// GetScList retrieves all SCdata records from the cache
// in arbitrary ket/value map order and returns a list
func GetFullSCDataList() ([]*SCData, error) {
	dbase, err := db.Get()
	if err != nil {
		return nil, err
	}
	ret := make([]*SCData, 0)
	err = dbase.ForEachPrefix(dbSCDataGroupPrefix(), func(entry database.Entry) bool {
		scdata := &SCData{}
		if err = json.Unmarshal(entry.Value, scdata); err == nil {
			ret = append(ret, scdata)
		}
		return false
	})
	return ret, nil
}

// prefix of the SCData entry key
func dbSCDataGroupPrefix() []byte {
	return []byte("sc_")
}

// key of the entry
func dbSCDataKey(scid *transaction.ScId) []byte {
	var buf bytes.Buffer
	buf.Write(dbSCDataGroupPrefix())
	buf.Write(scid.Bytes())
	return buf.Bytes()
}

// SaveSCData saves SCData record to the registry
// overwrites previous if any
// for new sc
func SaveSCData(scd *SCData) error {
	dbase, err := db.Get()
	if err != nil {
		return err
	}
	jsonData, err := json.Marshal(scd)
	if err != nil {
		return err
	}
	return dbase.Set(database.Entry{
		Key:   dbSCDataKey(scd.ScId),
		Value: jsonData,
	})
}
