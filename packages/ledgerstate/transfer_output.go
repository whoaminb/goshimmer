package ledgerstate

import (
	"encoding/binary"

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
	balanceCount := len(transferOutput.balances)
	coloredBalanceLength := colorLength + 64

	result := make([]byte, realityIdLength+balanceCount*coloredBalanceLength)

	copy(result[0:], transferOutput.realityId[:])

	binary.LittleEndian.PutUint32(result[realityIdLength:], uint32(balanceCount))
	for i := 0; i < balanceCount; i++ {
		copy(result[realityIdLength+4+i*coloredBalanceLength:], transferOutput.balances[i].color[:colorLength])
		binary.LittleEndian.PutUint64(result[realityIdLength+4+i*coloredBalanceLength+colorLength:], transferOutput.balances[i].balance)
	}

	return result, nil
}

func (transferOutput *TransferOutput) Unmarshal(key []byte, serializedObject []byte) (objectstorage.StorableObject, error) {
	result := &TransferOutput{
		id:       key,
		balances: transferOutput.unmarshalBalances(serializedObject[realityIdLength:]),
	}

	copy(result.transferHash[:], key[:transferHashLength])
	copy(result.addressHash[:], key[transferHashLength:transferHashLength+addressHashLength])
	copy(result.realityId[:], serializedObject[:realityIdLength])

	return result, nil
}

func (transferOutput *TransferOutput) unmarshalBalances(serializedBalances []byte) []*ColoredBalance {
	balanceCount := int(binary.LittleEndian.Uint32(serializedBalances))
	coloredBalanceLength := colorLength + 64

	balances := make([]*ColoredBalance, balanceCount)
	for i := 0; i < balanceCount; i++ {
		color := Color{}
		copy(color[:], serializedBalances[4+i*coloredBalanceLength:])

		balances[i] = NewColoredBalance(color, binary.LittleEndian.Uint64(serializedBalances[4+i*coloredBalanceLength+colorLength:]))
	}

	return balances
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
