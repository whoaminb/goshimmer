package waspconn

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/daemon"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/netutil/buffconn"
	"net"
)

type WaspConnector struct {
	id                             string
	bconn                          *buffconn.BufferedConnection
	subscriptions                  map[address.Address]int
	inTxChan                       chan *transaction.Transaction
	exitConnChan                   chan struct{}
	receiveValueTransactionClosure *events.Closure
	receiveWaspMessageClosure      *events.Closure
	closeClosure                   *events.Closure
	log                            *logger.Logger
}

func Run(conn net.Conn, log *logger.Logger) {
	wconn := &WaspConnector{
		id:           "wasp_" + conn.RemoteAddr().String(),
		bconn:        buffconn.NewBufferedConnection(conn),
		exitConnChan: make(chan struct{}),
		log:          log.Named(conn.RemoteAddr().String()),
	}
	err := daemon.BackgroundWorker(wconn.id, func(shutdownSignal <-chan struct{}) {
		select {
		case <-shutdownSignal:

			wconn.log.Infof("shutdown")

		case <-wconn.exitConnChan:
			wconn.log.Infof("closing..")
		}

		go wconn.detach()
	})

	if err != nil {
		close(wconn.exitConnChan)
		wconn.log.Errorf("can't start a deamon")
		return
	}
	wconn.attach()
}

func (wconn *WaspConnector) attach() {
	wconn.subscriptions = make(map[address.Address]int)
	wconn.inTxChan = make(chan *transaction.Transaction)

	wconn.receiveValueTransactionClosure = events.NewClosure(func(vtx *transaction.Transaction) {
		wconn.inTxChan <- vtx
	})

	wconn.receiveWaspMessageClosure = events.NewClosure(func(data []byte) {
		wconn.processMsgDataFromWasp(data)
	})

	wconn.closeClosure = events.NewClosure(func() {
		wconn.log.Info("Wasp connection closed")
	})

	// attach connector to the flow of incoming value transactions
	EventValueTransactionReceived.Attach(wconn.receiveValueTransactionClosure)

	wconn.bconn.Events.ReceiveMessage.Attach(wconn.receiveWaspMessageClosure)
	wconn.bconn.Events.Close.Attach(wconn.closeClosure)

	// read connection thread
	go func() {
		if err := wconn.bconn.Read(); err != nil {
			wconn.log.Errorf("error while reading socket: %v", err)
		}
		close(wconn.exitConnChan)
	}()

	// read incoming pre-filtered transactions from node
	go func() {
		for vtx := range wconn.inTxChan {
			wconn.processTransactionFromNode(vtx)
		}
	}()
}

func (wconn *WaspConnector) detach() {
	EventValueTransactionReceived.Detach(wconn.receiveValueTransactionClosure)
	wconn.bconn.Events.ReceiveMessage.Detach(wconn.receiveWaspMessageClosure)
	wconn.bconn.Events.Close.Detach(wconn.closeClosure)

	close(wconn.inTxChan)
	_ = wconn.bconn.Close()

	wconn.log.Debugf("detached waspconn")
}

func (wconn *WaspConnector) subscribe(addr *address.Address) {
	_, ok := wconn.subscriptions[*addr]
	if !ok {
		wconn.subscriptions[*addr] = 0
	}
}

func (wconn *WaspConnector) isSubscribed(addr *address.Address) bool {
	_, ok := wconn.subscriptions[*addr]
	return ok
}

// process parsed SC transaction incoming from the node.
// Forward to wasp if subscribed
func (wconn *WaspConnector) processTransactionFromNode(vtx *transaction.Transaction) {
	// determine if transaction contains any of subscribed addresses in its outputs

	isSubscribed := false
	vtx.Outputs().ForEach(func(addr address.Address, _ []*balance.Balance) bool {
		if wconn.isSubscribed(&addr) {
			isSubscribed = true
			return false
		}
		return true
	})
	if !isSubscribed {
		// dismiss unsubscribe transaction
		return
	}
	if err := wconn.sendTransactionToWasp(vtx); err != nil {
		wconn.log.Errorf("failed to send transaction")
	}
}

// find transaction async, parse it to SCTransaction and send to Wasp
func (wconn *WaspConnector) getTransaction(txid *transaction.ID) {
	// TODO
}

func (wconn *WaspConnector) getAddressBalance(addr *address.Address) {
	// TODO
}
