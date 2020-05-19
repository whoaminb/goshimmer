package utxodb

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address/signaturescheme"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/mr-tron/base58"
	"sync"
)

const (
	genesisPrivateKeyStr = "1pK9KraR4YTSHh3bq7hrigFSyq4HgWufhRyME84DPbwWpcoF1zwq6J1zaeyYUb8ut6ia9uQ9B9ughrpj2aZ7CMU"
	genesisPublicKeyStr  = "4f1W4o6PBKXXFsMHRcndYabpmaXPTdNNmtnV2NkfHnAa"
	supply               = int64(10 * 1000 * 1000 * 1000)
)

var (
	transactions     = make(map[transaction.ID]*transaction.Transaction)
	utxo             = make(map[transaction.OutputID]bool)
	utxoByAddress    = make(map[address.Address][]transaction.ID)
	mutexdb          sync.RWMutex
	genesisSigScheme signaturescheme.SignatureScheme
	genesisAddress   address.Address
)

func init() {
	var privKey ed25519.PrivateKey
	var pubKey ed25519.PublicKey
	var err error
	priv, err := base58.Decode(genesisPrivateKeyStr)
	if err != nil || len(priv) != len(privKey) {
		panic(err)
	}
	pub, err := base58.Decode(genesisPublicKeyStr)
	if err != nil || len(pub) != len(pubKey) {
		panic(err)
	}
	copy(privKey[:], priv)
	copy(pubKey[:], pub)
	genesisKeyPair := ed25519.KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
	}
	genesisSigScheme = signaturescheme.ED25519(genesisKeyPair)
	genesisAddress = genesisSigScheme.Address()

	// create genesis

	var niloutid transaction.OutputID

	inputs := transaction.NewInputs(niloutid)
	outputs := transaction.NewOutputs(map[address.Address][]*balance.Balance{
		genesisAddress: {balance.New(balance.ColorIOTA, supply)},
	})
	genesisTx := transaction.New(inputs, outputs)
	genesisTx.Sign(genesisSigScheme)
	transactions[genesisTx.ID()] = genesisTx
	utxoByAddress[genesisAddress] = []transaction.ID{genesisTx.ID()}
}

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

	// TODO check tx balance

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
		return false
	})

	tx.Outputs().ForEach(func(address address.Address, balances []*balance.Balance) bool {
		utxo[transaction.NewOutputID(address, tx.ID())] = true
		lst, ok := utxoByAddress[address]
		if !ok {
			lst = make([]transaction.ID, 0)
		}
		lst = append(lst, tx.ID())
		utxoByAddress[address] = lst
		return false
	})
	transactions[tx.ID()] = tx
	return nil
}

func GetTransaction(id transaction.ID) (*transaction.Transaction, bool) {
	mutexdb.RLock()
	defer mutexdb.RUnlock()

	tx, ok := transactions[id]
	return tx, ok
}

func MustGetTransaction(id transaction.ID) *transaction.Transaction {
	tx, ok := GetTransaction(id)
	if !ok {
		panic(fmt.Sprintf("tx id doesn't exist: %s", id.String()))
	}
	return tx
}

func GetAddressOutputs(addr address.Address) map[transaction.OutputID][]*balance.Balance {
	mutexdb.RLock()
	defer mutexdb.RUnlock()

	ret := make(map[transaction.OutputID][]*balance.Balance)

	txIds, ok := utxoByAddress[addr]
	if !ok || len(txIds) == 0 {
		return nil
	}
	for _, txid := range txIds {
		txInp := MustGetTransaction(txid)
		bals, ok := txInp.Outputs().Get(addr)
		if !ok {
			panic("output does not exist")
		}
		ret[transaction.NewOutputID(addr, txid)] = bals.([]*balance.Balance)
	}
	return ret
}

func TakeIotasFromGenesis(amount int64, targetAddress address.Address) (*transaction.Transaction, error) {
	genesisOutputs := GetAddressOutputs(genesisAddress)
	oids := make([]transaction.OutputID, 0)
	sum := int64(0)
	for oid, bals := range genesisOutputs {
		containsIotas := false
		for _, b := range bals {
			if b.Color() == balance.ColorIOTA {
				sum += b.Value()
				containsIotas = true
			}
		}
		if containsIotas {
			oids = append(oids, oid)
		}
		if sum >= amount {
			break
		}
	}
	inputs := transaction.NewInputs(oids...)

	out := map[address.Address][]*balance.Balance{targetAddress: {balance.New(balance.ColorIOTA, amount)}}
	if sum > amount {
		out[genesisAddress] = []*balance.Balance{balance.New(balance.ColorIOTA, sum-amount)}
	}

	outputs := transaction.NewOutputs(out)

	tx := transaction.New(inputs, outputs)
	if !checkInputsOutputs(tx) {
		panic("something wrong with inouts/outputs")
	}

	tx.Sign(genesisSigScheme)

	if !tx.SignaturesValid() {
		panic("something wrong with signatures")
	}
	return tx, nil
}
