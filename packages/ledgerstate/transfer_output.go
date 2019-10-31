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

	storageKey  []byte
	ledgerState *LedgerState
}

func NewTransferOutput(ledgerState *LedgerState, realityId RealityId, transferHash TransferHash, addressHash AddressHash, balances ...*ColoredBalance) *TransferOutput {
	return &TransferOutput{
		transferHash: transferHash,
		addressHash:  addressHash,
		balances:     balances,
		realityId:    realityId,

		storageKey:  append(transferHash[:], addressHash[:]...),
		ledgerState: ledgerState,
	}
}

func (transferOutput *TransferOutput) GetRealityId() RealityId {
	return transferOutput.realityId
}

func (transferOutput *TransferOutput) GetBalances() []*ColoredBalance {
	return transferOutput.balances
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

func (transferOutput *TransferOutput) GetStorageKey() []byte {
	return transferOutput.storageKey
}

func (transferOutput *TransferOutput) Update(other objectstorage.StorableObject) {}

func (transferOutput *TransferOutput) MarshalBinary() ([]byte, error) {
	balanceCount := len(transferOutput.balances)

	result := make([]byte, realityIdLength+4+balanceCount*coloredBalanceLength)

	copy(result[0:], transferOutput.realityId[:])

	binary.LittleEndian.PutUint32(result[realityIdLength:], uint32(balanceCount))
	for i := 0; i < balanceCount; i++ {
		copy(result[realityIdLength+4+i*coloredBalanceLength:], transferOutput.balances[i].color[:colorLength])
		binary.LittleEndian.PutUint64(result[realityIdLength+4+i*coloredBalanceLength+colorLength:], transferOutput.balances[i].balance)
	}

	return result, nil
}

func (transferOutput *TransferOutput) UnmarshalBinary(serializedObject []byte) error {
	if err := transferOutput.transferHash.UnmarshalBinary(transferOutput.storageKey[:transferHashLength]); err != nil {
		return err
	}

	if err := transferOutput.addressHash.UnmarshalBinary(transferOutput.storageKey[transferHashLength:]); err != nil {
		return err
	}

	if err := transferOutput.realityId.UnmarshalBinary(serializedObject[:realityIdLength]); err != nil {
		return err
	}

	if balances, err := transferOutput.unmarshalBalances(serializedObject[realityIdLength:]); err != nil {
		return err
	} else {
		transferOutput.balances = balances
	}

	return nil
}

func (transferOutput *TransferOutput) unmarshalBalances(serializedBalances []byte) ([]*ColoredBalance, error) {
	balanceCount := int(binary.LittleEndian.Uint32(serializedBalances))

	balances := make([]*ColoredBalance, balanceCount)
	for i := 0; i < balanceCount; i++ {
		coloredBalance := ColoredBalance{}
		if err := coloredBalance.UnmarshalBinary(serializedBalances[4+i*coloredBalanceLength:]); err != nil {
			return nil, err
		}

		balances[i] = &coloredBalance
	}

	return balances, nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
