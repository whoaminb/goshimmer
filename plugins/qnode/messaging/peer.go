package messaging

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/hive.go/backoff"
	"net"
	"sync"
	"time"
)

// represents point-to-point TCP connection between this qnode and another
// it is used as transport for message exchange
// Another end is always using the same connection
// the qnodePeer is used to send heartbeat messages and keeps last several of them
// in array. It is used to calculated average latency (network delay) between two local clocks
type qnodePeer struct {
	*sync.RWMutex
	peerconn     *peeredConnection // nil means not connected
	handshakeOk  bool
	peerPortAddr *registry.PortAddr
	startOnce    *sync.Once
	// heartbeats and latencies
	lastHeartbeatReceived time.Time
	lastHeartbeatSent     time.Time
	latency               [numHeartbeatsToKeep]int64
	hbIdx                 int
}

const (
	// equal and larger msg types are committee messages
	// those with smaller are reserved by the package for heartbeat and handshake messages
	FirstCommitteeMsgCode = byte(0x10)

	MsgTypeHeartbeat = byte(0)
	MsgTypeHandshake = byte(1)

	restartAfter = 1 * time.Second
	dialTimeout  = 1 * time.Second
	dialRetries  = 10
	backoffDelay = 500 * time.Millisecond
)

// retry net.Dial once, on fail after 0.5s
var dialRetryPolicy = backoff.ConstantBackOff(backoffDelay).With(backoff.MaxRetries(dialRetries))

func (c *qnodePeer) isInbound() bool {
	return isInboundAddr(c.peerPortAddr.String())
}

func (c *qnodePeer) connStatus() (bool, bool) {
	c.RLock()
	defer c.RUnlock()
	return c.peerconn != nil, c.handshakeOk
}

func (c *qnodePeer) closeConn() {
	c.Lock()
	defer c.Unlock()
	if c.peerconn != nil {
		_ = c.peerconn.Close()
	}
}

// dials outbound address and established connection
func (c *qnodePeer) runOutbound() {
	if c.isInbound() {
		return
	}
	if c.peerconn != nil {
		panic("c.peerconn != nil")
	}
	log.Debugf("runOutbound %s", c.peerPortAddr.String())

	defer c.runAfter(restartAfter)

	var conn net.Conn
	addr := fmt.Sprintf("%s:%d", c.peerPortAddr.Addr, c.peerPortAddr.Port)
	if err := backoff.Retry(dialRetryPolicy, func() error {
		var err error
		conn, err = net.DialTimeout("tcp", addr, dialTimeout)
		if err != nil {
			return fmt.Errorf("dial %s failed: %w", addr, err)
		}
		return nil
	}); err != nil {
		log.Error(err)
		return
	}
	//manconn := network.NewManagedConnection(conn)
	c.peerconn = newPeeredConnection(conn, c)
	if err := c.sendHandshake(); err != nil {
		log.Errorf("error during sendHandshake: %v", err)
		return
	}
	log.Debugf("starting reading outbound %s", c.peerPortAddr.String())
	if err := c.peerconn.Read(); err != nil {
		log.Error(err)
	}
	log.Debugf("stopped reading. Closing %s", c.peerPortAddr.String())
	c.closeConn()
}

// sends handshake message. It contains IP address of this end.
// The address is used by another end for peering
func (c *qnodePeer) sendHandshake() error {
	data, _ := wrapPacket(&unwrappedPacket{
		msgType: MsgTypeHandshake,
		data:    []byte(OwnPortAddr().String()),
	})
	num, err := c.peerconn.Write(data)
	log.Debugf("sendHandshake %d bytes to %s", num, c.peerPortAddr.String())
	return err
}

// callback to process parsed message from the peer
func (c *qnodePeer) receiveData(packet *unwrappedPacket) {
	c.receiveHeartbeat(packet.ts)
	if packet.msgType == MsgTypeHeartbeat {
		// no need for further processing
		return
	}
	// it can't be handshake message, so it is committee message
	// find a target committee
	committee, ok := getCommittee(packet.scid)
	if !ok {
		log.Errorw("message for unexpected scontract",
			"from", c.peerPortAddr.String(),
			"scid", packet.scid.Short(),
			"senderIndex", packet.senderIndex,
			"msgType", packet.msgType,
		)
		return
	}
	if packet.senderIndex >= committee.operator.CommitteeSize() || packet.senderIndex == committee.operator.PeerIndex() {
		log.Errorw("wrong sender index", "from", c.peerPortAddr.String(), "senderIndex", packet.senderIndex)
		return
	}
	// forward the data to the callback function, provided by the operator whe registered
	committee.recvDataCallback(packet.senderIndex, packet.msgType, packet.data, time.Unix(0, packet.ts))
}

func (c *qnodePeer) sendData(data []byte) error {
	c.RLock()
	defer c.RUnlock()

	if c.peerconn == nil {
		return fmt.Errorf("error while sending data: connection with %s not established", c.peerPortAddr.String())
	}
	num, err := c.peerconn.Write(data)
	if num != len(data) {
		return fmt.Errorf("not all bytes written. err = %v", err)
	}
	go c.scheduleNexHeartbeat()
	return err
}
