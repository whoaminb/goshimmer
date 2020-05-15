package waspconn

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/packages/waspconn"
)

func (wconn *WaspConnector) sendTransactionToWasp(vtx *valuetransaction.Transaction) error {
	msg := waspconn.WaspRecvTransactionMsg{vtx}
	var buf bytes.Buffer
	buf.WriteByte(waspconn.WaspRecvTransactionCode)
	buf.Write(msg.Encode())

	_, err := wconn.bconn.Write(buf.Bytes())
	return err
}

func (wconn *WaspConnector) sendBalancesToWasp(address *address.Address, balances map[valuetransaction.ID][]*balance.Balance) error {
	msg := waspconn.WaspRecvBalancesMsg{
		Address:  address,
		Balances: balances,
	}
	var buf bytes.Buffer
	buf.WriteByte(waspconn.WaspRecvBalancesCode)
	buf.Write(msg.Encode())

	_, err := wconn.bconn.Write(buf.Bytes())
	return err

}
