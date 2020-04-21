package messaging

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/hashing"
	"sync"
	"time"
)

var (
	// all committees, indexed by scid, smart contracts ID.
	// one committee object for one smart contact
	committees     = make(map[hashing.HashValue]*CommitteeConn)
	committeeMutex = &sync.RWMutex{}
)

// represents committee of qnodes which run a particular smart contract
type CommitteeConn struct {
	// consensus operator object. Operator object handles messages between committee members
	// to run the consensus on a particular smart contract
	operator SCOperator
	// receive message callback. It is provided by tho operator when it registers itself
	// committee conn calls this function whenever message for this smart contract/operator arrives
	// ts is message timestamp set by the sneder's clock
	recvDataCallback func(senderIndex uint16, msgType byte, msgData []byte, ts time.Time)
	// peers which belong to the committee of the smart contract
	peers []*qnodePeer
}

// finds a committee by smart contract id
func getCommittee(scid *hashing.HashValue) (*CommitteeConn, bool) {
	committeeMutex.RLock()
	defer committeeMutex.RUnlock()

	cconn, ok := committees[*scid]
	if !ok {
		return nil, false
	}
	return cconn, true
}

// return operator of the smart contract
func GetOperator(scid *hashing.HashValue) (SCOperator, bool) {
	comm, ok := getCommittee(scid)
	if !ok {
		return nil, false
	}
	return comm.operator, true
}

// This function is called by the operator to register itself.
// It returns the committee object. Operator later uses this object to communicate with another peers in the committee
func RegisterNewOperator(op SCOperator, recvDataCallback func(senderIndex uint16, msgType byte, msgData []byte, ts time.Time)) *CommitteeConn {
	committeeMutex.Lock()
	defer committeeMutex.Unlock()

	if cconn, ok := committees[*op.SContractID()]; ok {
		return cconn
	}
	ret := &CommitteeConn{
		operator:         op,
		recvDataCallback: recvDataCallback,
		peers:            make([]*qnodePeer, len(op.PeerLocations())),
	}
	for i := range ret.peers {
		if i == int(op.PeerIndex()) {
			continue
		}
		ret.peers[i] = addPeer(op.PeerLocations()[i])
	}
	committees[*op.SContractID()] = ret
	return ret
}

// sends marshalled data of the message to specified peer
func (cconn *CommitteeConn) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if targetPeerIndex == cconn.operator.PeerIndex() || int(targetPeerIndex) >= len(cconn.peers) {
		return fmt.Errorf("attempt to send message to the wrong peer index")
	}
	if msgType < FirstCommitteeMsgCode {
		panic("reserved msg type")
	}

	peer := cconn.peers[targetPeerIndex]

	var wrapped []byte

	wrapped, ts := wrapPacket(&unwrappedPacket{
		msgType:     msgType,
		scid:        cconn.operator.SContractID(),
		senderIndex: cconn.operator.PeerIndex(),
		data:        msgData,
	})

	peer.Lock()
	peer.lastHeartbeatSent = ts
	peer.Unlock()

	err := peer.sendData(wrapped)
	return err
}

// sends marshalled data of the message to all peer (not to itself)
// returns number if successful sends and timestamp common for all messages
func (cconn *CommitteeConn) SendMsgToPeers(msgType byte, msgData []byte) (uint16, time.Time) {
	if msgType == FirstCommitteeMsgCode {
		panic("reserved msg type")
	}

	var wrapped []byte
	wrapped, ts := wrapPacket(&unwrappedPacket{
		msgType:     msgType,
		scid:        cconn.operator.SContractID(),
		senderIndex: cconn.operator.PeerIndex(),
		data:        msgData,
	})
	var ret uint16

	for i := uint16(0); i < cconn.operator.CommitteeSize(); i++ {
		if i == cconn.operator.PeerIndex() {
			continue
		}
		peer := cconn.peers[i]
		peer.Lock()
		peer.lastHeartbeatSent = ts
		peer.Unlock()

		if err := peer.sendData(wrapped); err == nil {
			ret++
		}
	}
	return ret, ts
}

// return if peer is alive. Used by the operator to determine current leader
func (cconn *CommitteeConn) IsAlivePeer(peerIndex uint16) bool {
	if int(peerIndex) >= len(cconn.peers) {
		return false
	}
	ret, _ := cconn.peers[peerIndex].isAlive()
	return ret
}
