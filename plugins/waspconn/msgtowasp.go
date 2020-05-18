package waspconn

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn"
)

func (wconn *WaspConnector) sendTransactionToWasp(vtx *transaction.Transaction) error {
	msg := waspconn.WaspRecvTransactionMsg{vtx}
	var buf bytes.Buffer
	if err := buf.WriteByte(waspconn.WaspRecvTransactionCode); err != nil {
		return err
	}
	if err := msg.Write(&buf); err != nil {
		return err
	}

	_, err := wconn.bconn.Write(buf.Bytes())
	return err
}

func (wconn *WaspConnector) sendBalancesToWasp(address *address.Address, balances map[transaction.ID][]*balance.Balance) error {
	msg := waspconn.WaspRecvBalancesMsg{
		Address:  address,
		Balances: balances,
	}
	var buf bytes.Buffer
	if err := buf.WriteByte(waspconn.WaspRecvBalancesCode); err != nil {
		return err
	}
	if err := msg.Write(&buf); err != nil {
		return err
	}
	_, err := wconn.bconn.Write(buf.Bytes())
	return err

}
