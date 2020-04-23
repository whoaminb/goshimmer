package sctransaction

import (
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"io"
)

// state block of the SC transaction. Represents SC state update
// previous state block can be determined by the chain transfer of the SC token in the UTXO part of the
// transaction
type StateBlock struct {
	// scid of the SC which is updated
	// scid contains balance.NEW_COLOR in the scid.Color field for the origin transaction
	scid *ScId
	// stata index is 0 for the origin transaction
	// consensus maintains incremental sequence of state indexes
	stateIndex uint32
	// timestamp of the transaction. 0 means transaction is not timestamped
	timestamp int64
	// requestId tx hash + requestId index which originated this state update
	// this reference makes requestId (inputs to state update) immutable part of the state update
	requestId *RequestId
	// TODO may be nil, in this case it is just a timestamped checkpoint
	// otherwise it references StateBody
	stateUpdateHash *hashing.HashValue
}

func NewStateBlock(scid *ScId, stateIndex uint32) *StateBlock {
	return &StateBlock{
		scid:       scid,
		stateIndex: stateIndex,
	}
}

// getters/setters

func (sb *StateBlock) ScId() *ScId {
	return sb.scid
}

func (sb *StateBlock) StateIndex() uint32 {
	return sb.stateIndex
}

func (sb *StateBlock) Timestamp() int64 {
	return sb.timestamp
}

func (sb *StateBlock) RequestId() *RequestId {
	return sb.requestId
}

func (sb *StateBlock) StateUpdateHash() *hashing.HashValue {
	return sb.stateUpdateHash
}

func (sb *StateBlock) WithTimestamp(ts int64) *StateBlock {
	sb.timestamp = ts
	return sb
}

func (sb *StateBlock) WithRequestId(reqId *RequestId) *StateBlock {
	sb.requestId = reqId
	return sb
}

func (sb *StateBlock) WithStateUpdateHash(h *hashing.HashValue) *StateBlock {
	sb.stateUpdateHash = h
	return sb
}

// encoding
// important: each block starts with 65 bytes of scid

func (sb *StateBlock) Write(w io.Writer) error {
	if err := sb.scid.Write(w); err != nil {
		return err
	}
	if err := util.WriteUint32(w, sb.stateIndex); err != nil {
		return err
	}
	if err := util.WriteUint64(w, uint64(sb.timestamp)); err != nil {
		return err
	}
	if err := sb.requestId.Write(w); err != nil {
		return err
	}
	var b byte
	if sb.stateUpdateHash == nil {
		b = 0
	} else {
		b = 0xFF
	}
	if err := util.WriteByte(w, b); err != nil {
		return err
	}
	if sb.stateUpdateHash != nil {
		if err := sb.stateUpdateHash.Write(w); err != nil {
			return err
		}
	}
	return nil
}

func (sb *StateBlock) Read(r io.Reader) error {
	scid := new(ScId)
	if err := scid.Read(r); err != nil {
		return err
	}
	var stateIndex uint32
	if err := util.ReadUint32(r, &stateIndex); err != nil {
		return err
	}
	var timestamp uint64
	if err := util.ReadUint64(r, &timestamp); err != nil {
		return err
	}
	reqId := new(RequestId)
	if err := reqId.Read(r); err != nil {
		return err
	}
	var stateUpdateHash *hashing.HashValue
	if b, err := util.ReadByte(r); err != nil {
		return err
	} else {
		if b != 0 {
			stateUpdateHash := new(hashing.HashValue)
			if err := stateUpdateHash.Read(r); err != nil {
				return err
			}
		}
	}
	sb.scid = scid
	sb.stateIndex = stateIndex
	sb.timestamp = int64(timestamp)
	sb.requestId = reqId
	sb.stateUpdateHash = stateUpdateHash
	return nil
}
