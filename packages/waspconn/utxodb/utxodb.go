package utxodb

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"sync"
)

var (
	transactions  = make(map[transaction.ID]*transaction.Transaction)
	utxo          = make(map[transaction.OutputID]bool)
	utxoByAddress = make(map[address.Address][]transaction.ID)
	mutexdb       sync.RWMutex
)

func AddTransaction(tx *transaction.Transaction) error {
	if !checkInputsOutputs(tx) {
		return fmt.Errorf("wrong balance between inputs and outputs")
	}
	if !tx.SignaturesValid() {
		return fmt.Errorf("invalid signature")
	}

	mutexdb.Lock()
	defer mutexdb.Unlock()

	if _, ok := transactions[tx.ID()]; ok {
		return fmt.Errorf("duplicate transaction")
	}

	var err error

	// check if outputs exist
	tx.Inputs().ForEach(func(outputId transaction.OutputID) bool {
		if _, ok := utxo[outputId]; !ok {
			err = fmt.Errorf("output doesn't exist")
			return true
		}
		return false
	})
	if err != nil {
		return fmt.Errorf("invalid or conflicting inputs: %v", err)
	}

	// delete inputs from utxo ledger
	tx.Inputs().ForEach(func(outputId transaction.OutputID) bool {
		delete(utxo, outputId)
		lst, ok := utxoByAddress[outputId.Address()]
		if ok {
			newLst := make([]transaction.ID, 0, len(lst))
			for _, txid := range lst {
				if txid != outputId.TransactionID() {
					newLst = append(newLst, txid)
				}
			}
			utxoByAddress[outputId.Address()] = newLst
		}
		return true
	})
	// add outputs to utxo ledger
	tx.Outputs().ForEach(func(address address.Address, balances []*balance.Balance) bool {
		utxo[transaction.NewOutputID(address, tx.ID())] = true
		lst, ok := utxoByAddress[address]
		if !ok {
			lst = make([]transaction.ID, 0)
		}
		lst = append(lst, tx.ID())
		utxoByAddress[address] = lst
		return true
	})
	transactions[tx.ID()] = tx
	checkLedgerBalance()
	return nil
}

func GetTransaction(id transaction.ID) (*transaction.Transaction, bool) {
	mutexdb.RLock()
	defer mutexdb.RUnlock()

	return getTransaction(id)
}

func getTransaction(id transaction.ID) (*transaction.Transaction, bool) {
	tx, ok := transactions[id]
	return tx, ok
}

func mustGetTransaction(id transaction.ID) *transaction.Transaction {
	tx, ok := transactions[id]
	if !ok {
		panic(fmt.Sprintf("tx id doesn't exist: %s", id.String()))
	}
	return tx
}

func MustGetTransaction(id transaction.ID) *transaction.Transaction {
	mutexdb.RLock()
	defer mutexdb.RUnlock()
	return mustGetTransaction(id)
}

func GetAddressOutputs(addr address.Address) map[transaction.OutputID][]*balance.Balance {
	mutexdb.RLock()
	defer mutexdb.RUnlock()

	return getAddressOutputs(addr)
}

func getAddressOutputs(addr address.Address) map[transaction.OutputID][]*balance.Balance {
	ret := make(map[transaction.OutputID][]*balance.Balance)

	txIds, ok := utxoByAddress[addr]
	if !ok || len(txIds) == 0 {
		return nil
	}
	for _, txid := range txIds {
		txInp := mustGetTransaction(txid)
		bals, ok := txInp.Outputs().Get(addr)
		if !ok {
			panic("output does not exist")
		}
		ret[transaction.NewOutputID(addr, txid)] = bals.([]*balance.Balance)
	}
	return ret
}

func getOutputTotal(outid transaction.OutputID) (int64, error) {
	tx, ok := getTransaction(outid.TransactionID())
	if !ok {
		return 0, errors.New("no such transaction")
	}
	btmp, ok := tx.Outputs().Get(outid.Address())
	if !ok {
		return 0, errors.New("no such output")
	}
	bals := btmp.([]*balance.Balance)
	sum := int64(0)
	for _, b := range bals {
		sum += b.Value()
	}
	return sum, nil
}

func checkLedgerBalance() {
	total := int64(0)
	for outp := range utxo {
		b, err := getOutputTotal(outp)
		if err != nil {
			panic("Wrong ledger balance: " + err.Error())
		}
		total += b
	}
	if total != GetSupply() {
		panic("wrong ledger balance")
	}
}

type AddressStats struct {
	Total      int64
	NumOutputs int
}

func GetLedgerStats() map[address.Address]AddressStats {
	mutexdb.RLock()
	defer mutexdb.RUnlock()

	ret := make(map[address.Address]AddressStats)
	for addr := range utxoByAddress {
		outputs := getAddressOutputs(addr)
		total := int64(0)
		for outp := range outputs {
			s, err := getOutputTotal(outp)
			if err != nil {
				panic(err)
			}
			total += s
		}
		ret[addr] = AddressStats{
			Total:      total,
			NumOutputs: len(outputs),
		}
	}
	return ret
}
