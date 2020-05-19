package waspconn

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"io"
)

func (wconn *WaspConnector) sendMsgToWasp(msg interface{ Write(io.Writer) error }) error {
	data, err := waspconn.EncodeMsg(msg)
	if err != nil {
		return err
	}
	_, err = wconn.bconn.Write(data)
	return err
}

func (wconn *WaspConnector) sendTransactionToWasp(vtx *transaction.Transaction) error {
	return wconn.sendMsgToWasp(&waspconn.WaspFromNodeTransactionMsg{vtx})
}

func (wconn *WaspConnector) sendBalancesToWasp(address *address.Address, balances map[transaction.ID][]*balance.Balance) error {
	return wconn.sendMsgToWasp(&waspconn.WaspFromNodeBalancesMsg{
		Address:  address,
		Balances: balances,
	})
}
