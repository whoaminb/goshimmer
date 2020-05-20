package utxodb

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
)

func TransferIotas(amount int64, source, target address.Address) (*transaction.Transaction, error) {
	sourceOutputs := GetAddressOutputs(source)
	oids := make([]transaction.OutputID, 0)
	sum := int64(0)
	for oid, bals := range sourceOutputs {
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
	if sum < amount {
		return nil, fmt.Errorf("amount is too big")
	}
	inputs := transaction.NewInputs(oids...)

	out := map[address.Address][]*balance.Balance{target: {balance.New(balance.ColorIOTA, amount)}}
	if sum > amount {
		out[GetGenesisAddress()] = []*balance.Balance{balance.New(balance.ColorIOTA, sum-amount)}
	}

	outputs := transaction.NewOutputs(out)

	tx := transaction.New(inputs, outputs)
	if !checkInputsOutputs(tx) {
		panic("something wrong with inputs/outputs")
	}

	tx.Sign(GetSigScheme(source))

	if !tx.SignaturesValid() {
		panic("something wrong with signatures")
	}
	return tx, nil
}
