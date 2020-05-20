package utxodb

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestBasic(t *testing.T) {
	genTx, ok := GetTransaction(genesisTxId)
	assert.Equal(t, ok, true)
	assert.Equal(t, genTx.ID(), genesisTxId)
}

func TestKnowAddresses(t *testing.T) {
	for i := 0; i < 4; i++ {
		sigS := GetSigScheme(GetAddress(i))
		t.Logf("#%d address: %s", i, sigS.Address().String())
	}
}

func TestGenesis(t *testing.T) {
	gout := GetAddressOutputs(GetGenesisSigScheme().Address())
	expectedOutputId := transaction.NewOutputID(GetGenesisSigScheme().Address(), GetGenesisTransaction().ID())
	genBals, ok := gout[expectedOutputId]
	assert.Equal(t, ok, true)
	assert.Equal(t, len(genBals), 1)
	genBal := genBals[0]
	assert.Equal(t, genBal.Color(), balance.ColorIOTA)
	assert.Equal(t, genBal.Value(), GetSupply())
}

func TestTransfer(t *testing.T) {
	_, err := TransferIotas(1000000, GetGenesisAddress(), GetAddress(1))
	assert.Equal(t, err, nil)
}

func TestTransferAndBook(t *testing.T) {
	tx, err := TransferIotas(1000000, GetGenesisAddress(), GetAddress(1))
	assert.Equal(t, err, nil)

	err = AddTransaction(tx)
	assert.Equal(t, err, nil)

	tx, err = TransferIotas(1000000, GetGenesisAddress(), GetAddress(2))
	assert.Equal(t, err, nil)

	err = AddTransaction(tx)
	assert.Equal(t, err, nil)

	tx, err = TransferIotas(1000000, GetGenesisAddress(), GetAddress(3))
	assert.Equal(t, err, nil)

	err = AddTransaction(tx)
	assert.Equal(t, err, nil)

	stats := GetLedgerStats()
	for addr, st := range stats {
		t.Logf("%s: balance %d, num outputs %d", addr.String(), st.Total, st.NumOutputs)
	}
}
