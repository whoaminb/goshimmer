package waspconn

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
)

func (wconn *WaspConnector) sendTransactionToWasp(vtx *valuetransaction.Transaction) error {
	msg := WaspRecvTransactionMsg{vtx}
	var buf bytes.Buffer
	buf.WriteByte(WaspRecvTransactionCode)
	buf.Write(msg.Encode())

	_, err := wconn.bconn.Write(buf.Bytes())
	return err
}

func (wconn *WaspConnector) sendBalancesToWasp(address *address.Address, balances map[valuetransaction.ID][]*balance.Balance) error {
	msg := WaspRecvBalancesMsg{
		Address:  address,
		Balances: balances,
	}
	var buf bytes.Buffer
	buf.WriteByte(WaspRecvBalancesCode)
	buf.Write(msg.Encode())

	_, err := wconn.bconn.Write(buf.Bytes())
	return err

}
