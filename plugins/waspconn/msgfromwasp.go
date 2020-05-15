package waspconn

// process messages received from the Wasp
func (wconn *WaspConnector) processMsgDataFromWasp(data []byte) {
	switch data[0] {

	case WaspSendTransactionCode:
		msg := WaspSendTransactionMsg{}
		if err := msg.Decode(data[1:]); err != nil {
			log.Errorf("wrong 'WaspSendTransactionMsg' message from %s", wconn.id)
			return
		}
		// TODO post value transaction to the tangle

	case WaspSendSubscribeCode:
		msg := WaspSendSubscribeMsg{}
		if err := msg.Decode(data[1:]); err != nil {
			log.Errorf("wrong 'WaspSendSubscribeMsg' message from %s", wconn.id)
			return
		}
		for _, addr := range msg.Addresses {
			wconn.subscribe(&addr)
		}

	case WaspSendGetTransactionCode:
		msg := WaspSendGetTransactionMsg{}
		if err := msg.Decode(data[1:]); err != nil {
			log.Errorf("wrong 'WaspSendGetBalancesMsg' message from %s", wconn.id)
			return
		}
		wconn.getTransaction(msg.TxId)

	case WaspSendGetBalancesCode:
		msg := WaspSendGetBalancesMsg{}
		if err := msg.Decode(data[1:]); err != nil {
			log.Errorf("wrong 'WaspSendGetBalancesMsg' message from %s", wconn.id)
			return
		}
		wconn.getAddressBalance(msg.Address)

	default:
		log.Errorf("wrong message code from Wasp: %d", data[0])
	}
}
