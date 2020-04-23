package messaging

import (
	"fmt"
	"github.com/iotaledger/goshimmer/plugins/config"
	"github.com/iotaledger/goshimmer/plugins/qnode/parameters"
	"github.com/iotaledger/goshimmer/plugins/qnode/registry"
	"github.com/iotaledger/hive.go/daemon"
	"net"
	"sync"
	"time"
)

var (
	// all qnode peers maintained by the node
	// map index is IP addr:port
	peers      = make(map[string]*qnodePeer)
	peersMutex = &sync.RWMutex{}
)

func Init() {
	initLogger()

	if err := daemon.BackgroundWorker("Qnode connectOutboundLoop", func(shutdownSignal <-chan struct{}) {
		log.Debugf("starting qnode peering...")

		// start loops which looks after the qnodePeers and maintains them connected
		go connectOutboundLoop()
		go connectInboundLoop()
		//go countConnectionsLoop() // helper for testing

		<-shutdownSignal

		closeAll()

		log.Debugf("stopped qnode communications...")
	}); err != nil {
		panic(err)
	}
}

// IP address and port of this node
func OwnPortAddr() *registry.PortAddr {
	return &registry.PortAddr{
		Port: config.Node.GetInt(parameters.QNODE_PORT),
		Addr: "127.0.0.1", // TODO for testing only
	}
}

func closeAll() {
	peersMutex.Lock()
	defer peersMutex.Unlock()

	for _, cconn := range peers {
		cconn.closeConn()
	}
}

// determines if the address of the peer inbound or outbound
// it is guaranteed that the peer on the other end will always get the opposite result
// It is determined by comparing address string with own address string
// panics if address equals to the own address
// This helps peers to know their role in the peer-to-peer connection
func isInboundAddr(addr string) bool {
	own := OwnPortAddr().String()
	if own == addr {
		// shouldn't come to this point due to checks before
		panic("can't be same PortAddr")
	}
	return addr < own
}

// adds new connection to the peer pool
// if it already exists, returns existing
// connection added to the pool is picked by loops which will try to establish connection
func addPeer(portAddr *registry.PortAddr) *qnodePeer {
	peersMutex.Lock()
	peersMutex.Unlock()

	addr := portAddr.String()
	if qconn, ok := peers[addr]; ok {
		return qconn
	}
	peers[addr] = &qnodePeer{
		RWMutex:      &sync.RWMutex{},
		peerPortAddr: portAddr,
		startOnce:    &sync.Once{},
	}
	log.Debugf("added new peer %s inbound = %v", addr, peers[addr].isInbound())
	return peers[addr]
}

// wait some time the rests peer to be connected by the loops
func (c *qnodePeer) runAfter(d time.Duration) {
	go func() {
		time.Sleep(d)
		c.Lock()
		c.startOnce = &sync.Once{}
		c.Unlock()
		log.Debugf("will run %s again", c.peerPortAddr.String())
	}()
}

// for testing
func countConnectionsLoop() {
	var totalNum, inboundNum, outboundNum, inConnectedNum, outConnectedNum, inHSNum, outHSNum int
	for {
		time.Sleep(2 * time.Second)
		totalNum, inboundNum, outboundNum, inConnectedNum, outConnectedNum, inHSNum, outHSNum = 0, 0, 0, 0, 0, 0, 0
		peersMutex.Lock()
		for _, c := range peers {
			totalNum++
			isConn, isHandshaken := c.connStatus()
			if c.isInbound() {
				inboundNum++
				if isConn {
					inConnectedNum++
				}
				if isHandshaken {
					inHSNum++
				}
			} else {
				outboundNum++
				if isConn {
					outConnectedNum++
				}
				if isHandshaken {
					outHSNum++
				}
			}
		}
		peersMutex.Unlock()
		log.Debugf("CONN STATUS: total conn: %d, in: %d, out: %d, inConnected: %d, outConnected: %d, inHS: %d, outHS: %d",
			totalNum, inboundNum, outboundNum, inConnectedNum, outConnectedNum, inHSNum, outHSNum)
	}
}

// loop which maintains outbound peers connected (if possible)
func connectOutboundLoop() {
	for {
		time.Sleep(100 * time.Millisecond)
		peersMutex.Lock()
		for _, c := range peers {
			c.startOnce.Do(func() {
				go c.runOutbound()
			})
		}
		peersMutex.Unlock()
	}
}

// loop which maintains inbound peers connected (when possible)
func connectInboundLoop() {
	listenOn := fmt.Sprintf(":%d", config.Node.GetInt(parameters.QNODE_PORT))
	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		log.Errorf("tcp listen on %s failed: %v. Restarting connectInboundLoop after 1 sec", listenOn, err)
		go func() {
			time.Sleep(1 * time.Second)
			connectInboundLoop()
		}()
		return
	}
	log.Infof("tcp listen inbound on %s", listenOn)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Errorf("failed accepting a connection request: %v", err)
			continue
		}
		log.Debugf("accepted connection from %s", conn.RemoteAddr().String())

		//manconn := network.NewManagedConnection(conn)
		bconn := newPeeredConnection(conn, nil)
		go func() {
			log.Debugf("starting reading inbound %s", conn.RemoteAddr().String())
			if err := bconn.Read(); err != nil {
				log.Error(err)
			}
			_ = bconn.Close()
		}()
	}
}
