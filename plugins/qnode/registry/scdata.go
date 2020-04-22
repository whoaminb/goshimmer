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

// GetScList retrieves all SCdata records from the registry
// in arbitrary key/value map order and returns a list
// if ownPortAddr is not nil, it only includes those SCData records which are processed
// by his node
func GetSCDataList(ownAddr *PortAddr) ([]*SCData, error) {
	dbase, err := db.Get()
	if err != nil {
		return nil, err
	}
	ret := make([]*SCData, 0)
	err = dbase.ForEachPrefix(dbSCDataGroupPrefix, func(entry database.Entry) bool {
		scdata := &SCData{}
		if err = json.Unmarshal(entry.Value, scdata); err == nil {
			if validate(scdata, ownAddr) {
				ret = append(ret, scdata)
			}
		}
		return false
	})
	return ret, err
}

// checks if SCData record is valid
// if ownAddr != nil checks if it is of interest to the current node
func validate(scdata *SCData, ownAddr *PortAddr) bool {
	addr := scdata.ScId.Address()
	dkshare, ok, _ := GetDKShare(&addr)
	if !ok {
		// failed to load dkshare of the sc address
		return false
	}
	if int(dkshare.Index) >= len(scdata.NodeLocations) {
		// shouldn't be
		return false
	}
	if ownAddr == nil {
		return true
	}
	if ownAddr.String() != scdata.NodeLocations[dkshare.Index].String() {
		// if own address is not consistent with the one at the index in the list of nodes
		// this node is not interested in the SC
		return false
	}
	return true
}

// prefix of the SCData entry key
var dbSCDataGroupPrefix = HashStrings("scdata").Bytes()

// key of the entry
func dbSCDataKey(scid *transaction.ScId) []byte {
	var buf bytes.Buffer
	buf.Write(dbSCDataGroupPrefix)
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
