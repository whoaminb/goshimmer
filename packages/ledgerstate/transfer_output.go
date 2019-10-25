package ledgerstate

import (
	"github.com/iotaledger/goshimmer/packages/objectstorage"
	"github.com/iotaledger/goshimmer/packages/stringify"
)

type TransferOutput struct {
	transferHash TransferHash
	addressHash  AddressHash
	balances     []*ColoredBalance
	realityId    RealityId

	id          []byte
	ledgerState *LedgerState
}

func NewTransferOutput(ledgerState *LedgerState, realityId RealityId, transferHash TransferHash, addressHash AddressHash, balances ...*ColoredBalance) *TransferOutput {
	return &TransferOutput{
		transferHash: transferHash,
		addressHash:  addressHash,
		balances:     balances,
		realityId:    realityId,

		id:          append(transferHash[:], addressHash[:]...),
		ledgerState: ledgerState,
	}
}

func (transferOutput *TransferOutput) String() string {
	return stringify.Struct("TransferOutput",
		stringify.StructField("transferHash", transferOutput.transferHash.String()),
		stringify.StructField("addressHash", transferOutput.addressHash.String()),
		stringify.StructField("balances", transferOutput.balances),
		stringify.StructField("realityId", transferOutput.realityId.String()),
	)
}

// region support object storage ///////////////////////////////////////////////////////////////////////////////////////

func (transferOutput *TransferOutput) GetId() []byte {
	return transferOutput.id
}

func (transferOutput *TransferOutput) Update(other objectstorage.StorableObject) {}

func (transferOutput *TransferOutput) Marshal() ([]byte, error) {
	return transferOutput.realityId[:], nil
}

func (transferOutput *TransferOutput) Unmarshal(key []byte, serializedObject []byte) (objectstorage.StorableObject, error) {
	result := &TransferOutput{
		id: key,
	}

	copy(result.transferHash[:], key[:transferHashLength])
	copy(result.addressHash[:], key[transferHashLength:transferHashLength+addressHashLength])
	copy(result.realityId[:], serializedObject)

	return result, nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
