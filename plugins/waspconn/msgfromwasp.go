package waspconn

import (
	"github.com/iotaledger/goshimmer/packages/waspconn"
)

// process messages received from the Wasp
func (wconn *WaspConnector) processMsgDataFromWasp(data []byte) {
	var msg interface{}
	var err error
	if msg, err = waspconn.DecodeMsg(data, false); err != nil {
		log.Errorf("DecodeMsg id %s, error: %v", wconn.id, err)
		return
	}
	switch msgt := msg.(type) {
	case *waspconn.WaspToNodeTransactionMsg:
		// TODO post value transaction to the tangle

	case *waspconn.WaspToNodeSubscribeMsg:
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
