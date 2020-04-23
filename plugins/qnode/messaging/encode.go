package messaging

import (
	"bytes"
	"fmt"
	. "github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"github.com/iotaledger/goshimmer/plugins/qnode/util"
	"time"
)

// structure of the message:
// timestamp   8 bytes
// msg type    1 byte
// is type != 0, next:
// scid 32 bytes
// sender index 2 bytes
// data variable bytes

type unwrappedPacket struct {
	ts          int64
	msgType     byte
	scid        *HashValue
	senderIndex uint16
	data        []byte
}

func unwrapPacket(data []byte) (*unwrappedPacket, error) {
	if len(data) < 9 {
		return nil, fmt.Errorf("too short message")
	}
	rdr := bytes.NewBuffer(data)
	var uts uint64
	err := util.ReadUint64(rdr, &uts)
	if err != nil {
		return nil, err
	}
	ret := &unwrappedPacket{
		ts: int64(uts),
	}
	ret.msgType, err = util.ReadByte(rdr)
	if err != nil {
		return nil, err
	}
	switch {
	case ret.msgType == MsgTypeHeartbeat:
		return ret, nil

	case ret.msgType == MsgTypeHandshake:
		ret.data = rdr.Bytes()
		return ret, nil

	case ret.msgType >= FirstCommitteeMsgCode:
		// committee message
		var scid HashValue
		_, err = rdr.Read(scid.Bytes())
		if err != nil {
			return nil, err
		}
		ret.scid = &scid
		err = util.ReadUint16(rdr, &ret.senderIndex)
		if err != nil {
			return nil, err
		}
		var dataLen uint32
		err = util.ReadUint32(rdr, &dataLen)
		if err != nil {
			return nil, err
		}
		ret.data = rdr.Bytes()
		if len(ret.data) != int(dataLen) {
			return nil, fmt.Errorf("unexpected data length")
		}
		return ret, nil
	}
	return nil, fmt.Errorf("wrong message type %d", ret.msgType)
}

// always puts timestamp into first 8 bytes and 1 byte msg type

func wrapPacket(up *unwrappedPacket) ([]byte, time.Time) {
	var buf bytes.Buffer
	// puts timestamp
	ts := time.Now()
	_ = util.WriteUint64(&buf, uint64(ts.UnixNano()))
	switch {
	case up == nil:
		buf.WriteByte(MsgTypeHeartbeat)

	case up.msgType == MsgTypeHeartbeat:
		buf.WriteByte(MsgTypeHeartbeat)

	case up.msgType == MsgTypeHandshake:
		buf.WriteByte(MsgTypeHandshake)
		buf.Write(up.data)

	case up.msgType >= FirstCommitteeMsgCode:
		buf.WriteByte(up.msgType)
		buf.Write(up.scid.Bytes())
		_ = util.WriteUint16(&buf, up.senderIndex)
		_ = util.WriteUint32(&buf, uint32(len(up.data)))
		buf.Write(up.data)

	default:
		log.Panicf("wrong msg type %d", up.msgType)
	}
	return buf.Bytes(), ts
}
