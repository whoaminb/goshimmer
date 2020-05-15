package waspconn

import "github.com/iotaledger/goshimmer/packages/waspconn"

// process messages received from the Wasp
func (wconn *WaspConnector) processMsgDataFromWasp(data []byte) {
	switch data[0] {

	case waspconn.WaspSendTransactionCode:
		msg := waspconn.WaspSendTransactionMsg{}
		if err := msg.Decode(data[1:]); err != nil {
			log.Errorf("wrong 'WaspSendTransactionMsg' message from %s", wconn.id)
			return
		}
		// TODO post value transaction to the tangle

	case waspconn.WaspSendSubscribeCode:
		msg := waspconn.WaspSendSubscribeMsg{}
		if err := msg.Decode(data[1:]); err != nil {
			log.Errorf("wrong 'WaspSendSubscribeMsg' message from %s", wconn.id)
			return
		}
		for _, addr := range msg.Addresses {
			wconn.subscribe(&addr)
		}

	case waspconn.WaspSendGetTransactionCode:
		msg := waspconn.WaspSendGetTransactionMsg{}
		if err := msg.Decode(data[1:]); err != nil {
			log.Errorf("wrong 'WaspSendGetBalancesMsg' message from %s", wconn.id)
			return
		}
		wconn.getTransaction(msg.TxId)

	case waspconn.WaspSendGetBalancesCode:
		msg := waspconn.WaspSendGetBalancesMsg{}
		if err := msg.Decode(data[1:]); err != nil {
			log.Errorf("wrong 'WaspSendGetBalancesMsg' message from %s", wconn.id)
			return
		}
		wconn.getAddressBalance(msg.Address)

	default:
		log.Errorf("wrong message code from Wasp: %d", data[0])
	}
}
