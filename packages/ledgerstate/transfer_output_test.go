package ledgerstate

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"

	"github.com/iotaledger/goshimmer/packages/binary/address"

	"github.com/magiconair/properties/assert"
)

func TestTransferOutput_MarshalUnmarshal(t *testing.T) {
	transferOutput := NewTransferOutput(nil, reality.NewId("REALITY"), transfer.NewHash("RECEIVE"), address.New([]byte("ADDRESS1")), NewColoredBalance(NewColor("IOTA"), 44), NewColoredBalance(NewColor("BTC"), 88))
	transferOutput.consumers = make(map[transfer.Hash][]address.Address)

	spendTransferHash := transfer.NewHash("SPEND")
	transferOutput.consumers[spendTransferHash] = make([]address.Address, 2)
	transferOutput.consumers[spendTransferHash][0] = address.New([]byte("ADDRESS2"))
	transferOutput.consumers[spendTransferHash][1] = address.New([]byte("ADDRESS3"))

	marshaledData, err := transferOutput.MarshalBinary()
	if err != nil {
		t.Error(err)
	}

	unmarshaledTransferOutput := TransferOutput{storageKey: transferOutput.GetStorageKey()}
	if err := unmarshaledTransferOutput.UnmarshalBinary(marshaledData); err != nil {
		t.Error(err)
	}

	assert.Equal(t, unmarshaledTransferOutput.realityId, transferOutput.realityId)
	assert.Equal(t, unmarshaledTransferOutput.transferHash, transferOutput.transferHash)
	assert.Equal(t, unmarshaledTransferOutput.addressHash, transferOutput.addressHash)
	assert.Equal(t, len(unmarshaledTransferOutput.consumers), len(transferOutput.consumers))
	assert.Equal(t, len(unmarshaledTransferOutput.consumers[spendTransferHash]), len(transferOutput.consumers[spendTransferHash]))
	assert.Equal(t, unmarshaledTransferOutput.consumers[spendTransferHash][0], transferOutput.consumers[spendTransferHash][0])
	assert.Equal(t, unmarshaledTransferOutput.consumers[spendTransferHash][1], transferOutput.consumers[spendTransferHash][1])
}
