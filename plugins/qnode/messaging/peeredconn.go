package messaging

import (
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/netutil/buffconn"
	"net"
)

// extension of BufferedConnection from hive.go
// BufferedConnection is a wrapper for net.Conn
// peeredConnection first handles handshake and then links
// with peer (peers) according to handshake information
type peeredConnection struct {
	*buffconn.BufferedConnection
	peer        *qnodePeer
	handshakeOk bool
}

// creates new peered connection and attach event handlers for received data and closing
func newPeeredConnection(conn net.Conn, peer *qnodePeer) *peeredConnection {
	bconn := &peeredConnection{
		BufferedConnection: buffconn.NewBufferedConnection(conn),
		peer:               peer,
	}
	bconn.Events.ReceiveMessage.Attach(events.NewClosure(func(data []byte) {
		bconn.receiveData(data)
	}))
	bconn.Events.Close.Attach(events.NewClosure(func() {
		if bconn.peer != nil {
			bconn.peer.Lock()
			bconn.peer.peerconn = nil
			bconn.peer.handshakeOk = false
			bconn.peer.Unlock()
		}
	}))
	return bconn
}

// receive data handler for peered connection
func (bconn *peeredConnection) receiveData(data []byte) {
	packet, err := unwrapPacket(data)
	if err != nil {
		log.Error("unwrapPacket: %v", err)
		bconn.peer.closeConn()
		return
	}
	if bconn.peer != nil {
		// it is peered but maybe not handshaked yet (can be only outbound)
		if bconn.peer.handshakeOk {
			// it is handshaked
			// just receive data
			bconn.peer.receiveData(packet)
		} else {
			// not handshaked => do handshake
			bconn.processHandShakeOutbound(packet)
		}
	} else {
		// not peered yet can be only inbound
		// peer up and do handshhake
		bconn.processHandShakeInbound(packet)
	}
}

// receives handshake response from the outbound peer
// assumes the connection is already peered (i can be only for outbound peers)

func (bconn *peeredConnection) processHandShakeOutbound(packet *unwrappedPacket) {
	if packet.msgType != MsgTypeHandshake {
		log.Errorf("unexpected message during handshake")
		return
	}
	peerAddr := string(packet.data)
	log.Debugf("received handshake from outbound %s", peerAddr)
	if peerAddr != bconn.peer.peerPortAddr.String() {
		log.Error("closeConn the peer connection: wrong handshake message from outbound peer: expected %s got '%s'",
			bconn.peer.peerPortAddr.String(), peerAddr)
		bconn.peer.closeConn()
	} else {
		log.Infof("handshake ok with peer %s", peerAddr)
		bconn.peer.handshakeOk = true

		bconn.peer.initHeartbeats()
		bconn.peer.receiveHeartbeat(packet.ts)
		go bconn.peer.scheduleNexHeartbeat()
	}
}

// receives handshake from the inbound peer
// links connection with the peer
// sends response back to finish the handshake
func (bconn *peeredConnection) processHandShakeInbound(packet *unwrappedPacket) {
	if packet.msgType != MsgTypeHandshake {
		log.Errorf("unexpected message during handshake")
		return
	}
	peerAddr := string(packet.data)
	log.Debugf("received handshake from inbound %s", peerAddr)

	peersMutex.RLock()
	peer, ok := peers[peerAddr]
	peersMutex.RUnlock()

	if !ok || !peer.isInbound() {
		log.Errorf("inbound connection from unexpected peer %s. Closing..", peerAddr)
		_ = bconn.Close()
		return
	}
	bconn.peer = peer

	peer.Lock()
	peer.peerconn = bconn
	peer.handshakeOk = true
	peer.Unlock()

	if err := peer.sendHandshake(); err == nil {
		bconn.peer.initHeartbeats()
		bconn.peer.receiveHeartbeat(packet.ts)
		go bconn.peer.scheduleNexHeartbeat()
	} else {
		log.Error("error while responding to handshake: %v. Closing connection", err)
		_ = bconn.Close()
	}

}
