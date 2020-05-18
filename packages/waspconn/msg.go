package waspconn

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"io"
)

const (
	// wasp -> node
	WaspSendTransactionCode    = byte(1)
	WaspSendSubscribeCode      = byte(2)
	WaspSendGetTransactionCode = byte(3)
	WaspSendGetBalancesCode    = byte(4)

	// node -> wasp
	WaspRecvTransactionCode = byte(5)
	WaspRecvBalancesCode    = byte(6)
)

type WaspSendTransactionMsg struct {
	Tx *transaction.Transaction
}

type WaspSendSubscribeMsg struct {
	Addresses   []address.Address
	PullBacklog bool
}

type WaspSendGetTransactionMsg struct {
	TxId *transaction.ID
}

type WaspSendGetBalancesMsg struct {
	Address *address.Address
}

type WaspRecvTransactionMsg struct {
	Tx *transaction.Transaction
}

type WaspRecvBalancesMsg struct {
	Address  *address.Address
	Balances map[transaction.ID][]*balance.Balance
}

func (msg *WaspSendTransactionMsg) Write(w io.Writer) error {
	return WriteBytes32(w, msg.Tx.Bytes())
}

func (msg *WaspSendTransactionMsg) Read(r io.Reader) error {
	var err error
	data, err := ReadBytes32(r)
	if err != nil {
		return err
	}
	msg.Tx, _, err = transaction.FromBytes(data)
	return err
}

func (msg *WaspSendSubscribeMsg) Write(w io.Writer) error {
	if err := WriteUint16(w, uint16(len(msg.Addresses))); err != nil {
		return err
	}
	for _, addr := range msg.Addresses {
		if _, err := w.Write(addr.Bytes()); err != nil {
			return err
		}
	}
	return WriteBoolByte(w, msg.PullBacklog)
}

func (msg *WaspSendSubscribeMsg) Read(r io.Reader) error {
	var size uint16
	if err := ReadUint16(r, &size); err != nil {
		return err
	}
	msg.Addresses = make([]address.Address, size)
	for i := range msg.Addresses {
		n, err := r.Read(msg.Addresses[i][:])
		if err != nil {
			return err
		}
		if n != balance.ColorLength {
			return fmt.Errorf("error while reading 'subscribe' message")
		}
	}
	if err := ReadBoolByte(r, &msg.PullBacklog); err != nil {
		return err
	}
	return nil
}

func (msg *WaspSendGetTransactionMsg) Write(w io.Writer) error {
	_, err := w.Write(msg.TxId.Bytes())
	return err
}

func (msg *WaspSendGetTransactionMsg) Read(r io.Reader) error {
	msg.TxId = new(transaction.ID)
	n, err := r.Read(msg.TxId[:])
	if err != nil {
		return err
	}
	if n != transaction.IDLength {
		return fmt.Errorf("error while reading 'get transaction' message")
	}
	return nil
}

func (msg *WaspSendGetBalancesMsg) Write(w io.Writer) error {
	_, err := w.Write(msg.Address.Bytes())
	return err
}

func (msg *WaspSendGetBalancesMsg) Read(r io.Reader) error {
	msg.Address = new(address.Address)
	n, err := r.Read(msg.Address[:])
	if err != nil {
		return err
	}
	if n != address.Length {
		return fmt.Errorf("error while reading 'get balances' message")
	}
	return nil
}

func (msg *WaspRecvTransactionMsg) Write(w io.Writer) error {
	return WriteBytes32(w, msg.Tx.Bytes())
}

func (msg *WaspRecvTransactionMsg) Read(r io.Reader) error {
	data, err := ReadBytes32(r)
	if err != nil {
		return err
	}
	msg.Tx, _, err = transaction.FromBytes(data)
	return err
}

func (msg *WaspRecvBalancesMsg) Write(w io.Writer) error {
	_, err := w.Write(msg.Address.Bytes())
	if err != nil {
		return err
	}
	return WriteBalances(w, msg.Balances)
}

func (msg *WaspRecvBalancesMsg) Read(r io.Reader) error {
	msg.Address = new(address.Address)
	n, err := r.Read(msg.Address[:])
	if err != nil {
		return err
	}
	if n != address.Length {
		return fmt.Errorf("error while decoding 'recv balance' message")
	}
	if msg.Balances, err = ReadBalances(r); err != nil {
		return err
	}
	return nil
}

func WriteBalances(w io.Writer, balances map[transaction.ID][]*balance.Balance) error {
	if err := WriteUint16(w, uint16(len(balances))); err != nil {
		return err
	}
	for txid, bals := range balances {
		if _, err := w.Write(txid[:]); err != nil {
			return err
		}
		if err := WriteUint16(w, uint16(len(bals))); err != nil {
			return err
		}
		for _, b := range bals {
			if _, err := w.Write(b.Color().Bytes()); err != nil {
				return err
			}
			if err := WriteUint64(w, uint64(b.Value())); err != nil {
				return err
			}
		}
	}
	return nil
}

func ReadBalances(r io.Reader) (map[transaction.ID][]*balance.Balance, error) {
	var size uint16
	if err := ReadUint16(r, &size); err != nil {
		return nil, err
	}
	ret := make(map[transaction.ID][]*balance.Balance, size)
	for i := uint16(0); i < size; i++ {
		txid := new(transaction.ID)
		n, err := r.Read(txid.Bytes())
		if err != nil {
			return nil, err
		}
		if n != transaction.IDLength {
			return nil, fmt.Errorf("error while decoding 'recv balance' message")
		}
		var numBals uint16
		if err := ReadUint16(r, &numBals); err != nil {
			return nil, err
		}
		lst := make([]*balance.Balance, numBals)
		for i := range lst {
			var color balance.Color
			n, err := r.Read(color[:])
			if err != nil {
				return nil, err
			}
			if n != balance.ColorLength {
				return nil, fmt.Errorf("error while decoding 'recv balance' message")
			}
			var value uint64
			if err := ReadUint64(r, &value); err != nil {
				return nil, err
			}
			lst[i] = balance.New(color, int64(value))
		}
		ret[*txid] = lst
	}
	return ret, nil
}
