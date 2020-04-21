package messaging

import (
	"time"
)

// message type is 1 byte
// from 0 until maxSpecMsgCode inclusive it is reserved for heartbeat and other message types
// these messages are processed by processHeartbeat method
// the rest are forwarded to SC operator

const (
	// heartbeat msg
	numHeartbeatsToKeep = 5               // number of heartbeats to save for average latency
	heartbeatEvery      = 5 * time.Second // heartBeat period
	isDeadAfterMissing  = 2               // is dead after 4 heartbeat periods missing
)

func (c *qnodePeer) initHeartbeats() {
	c.lastHeartbeatSent = time.Time{}
	c.lastHeartbeatReceived = time.Time{}
	c.hbIdx = 0
	for i := range c.latency {
		c.latency[i] = 0
	}
}

func (c *qnodePeer) receiveHeartbeat(ts int64) {
	c.Lock()
	c.lastHeartbeatReceived = time.Now()
	lagNano := c.lastHeartbeatReceived.UnixNano() - ts
	c.latency[c.hbIdx] = lagNano
	c.hbIdx = (c.hbIdx + 1) % numHeartbeatsToKeep
	c.Unlock()

	//log.Debugf("heartbeat received from %s, lag %f milisec", c.peerPortAddr.String(), float64(lagNano/10000)/100)
}

func (c *qnodePeer) scheduleNexHeartbeat() {
	time.Sleep(heartbeatEvery)
	if peerAlive, _ := c.isAlive(); !peerAlive {
		log.Debugf("stopped sending heartbeat: peer %s is dead", c.peerPortAddr.String())
		return
	}

	c.Lock()

	if time.Since(c.lastHeartbeatSent) < heartbeatEvery {
		// was recently sent. exit
		c.Unlock()
		return
	}
	var hbMsgData []byte
	hbMsgData, c.lastHeartbeatSent = wrapPacket(nil)

	c.Unlock()

	_ = c.sendData(hbMsgData)
	//log.Debugf("sent heartbeat to %s", c.peerPortAddr.String())

	// repeat after some time
}

// return true if is alive and average latency in nanosec
func (c *qnodePeer) isAlive() (bool, int64) {
	c.RLock()
	defer c.RUnlock()
	if c.peerconn == nil || !c.handshakeOk {
		return false, 0
	}

	if time.Since(c.lastHeartbeatReceived) > heartbeatEvery*isDeadAfterMissing {
		return false, 0
	}
	sum := int64(0)
	for _, l := range c.latency {
		sum += l
	}
	return true, sum / numHeartbeatsToKeep
}
