package waspconn

import (
	"github.com/iotaledger/goshimmer/packages/waspconn"
)

// process messages received from the Wasp
func (wconn *WaspConnector) processMsgDataFromWasp(data []byte) {
	var msg interface{}
	var err error
	if msg, err = waspconn.DecodeMsg(data, false); err != nil {
		wconn.log.Errorf("DecodeMsg: %v", err)
		return
	}
	switch msgt := msg.(type) {
	case *waspconn.WaspPingMsg:
		wconn.log.Debugf("PING %d received", msgt.Id)
		if err := wconn.sendMsgToWasp(msgt); err != nil {
			wconn.log.Errorf("responding to ping: %v", err)
		}

	case *waspconn.WaspToNodeTransactionMsg:
		wconn.log.Debugf("'WaspToNodeTransactionMsg' received: txid %s", msgt.Tx.ID().String())
		wconn.postTransaction(msgt.Tx)

	case *waspconn.WaspToNodeSubscribeMsg:
		wconn.log.Debugf("WaspToNodeSubscribeMsg: %+v", msgt.Addresses)
		for _, addr := range msgt.Addresses {
			wconn.subscribe(&addr)
		}

	case *waspconn.WaspToNodeGetTransactionMsg:
		wconn.getTransaction(msgt.TxId)

	case *waspconn.WaspToNodeGetBalancesMsg:
		wconn.getAddressBalance(msgt.Address)

	default:
		panic("wrong msg type")
	}
}
