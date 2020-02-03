package transferoutput

import (
	"encoding/binary"
	"sync"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/coloredcoins"
	"github.com/iotaledger/goshimmer/packages/stringify"
	"github.com/iotaledger/hive.go/objectstorage"
)

type TransferOutput struct {
	objectstorage.StorableObjectFlags

	id       Id
	spent    bool
	balances []*coloredcoins.ColoredBalance

	realityIdMutex sync.RWMutex
}

func New(id Id, balances ...*coloredcoins.ColoredBalance) *TransferOutput {
	return &TransferOutput{
		id:       id,
		balances: balances,
	}
}

func FromStorage(key []byte) objectstorage.StorableObject {
	result := &TransferOutput{}
	offset := 0

	if err := result.id.UnmarshalBinary(key[offset:]); err != nil {
		panic(err)
	}

	return result
}

func (transferOutput *TransferOutput) GetId() Id {
	return transferOutput.id
}

func (transferOutput *TransferOutput) GetBalances() []*coloredcoins.ColoredBalance {
	return transferOutput.balances
}

func (transferOutput *TransferOutput) IsSpent() (result bool) {
	// TODO: IMPLEMENT

	return
}

func (transferOutput *TransferOutput) String() string {
	return stringify.Struct("TransferOutput",
		stringify.StructField("id", transferOutput.GetId().String()),
		stringify.StructField("balances", transferOutput.GetBalances()),
		// TODO: IS SPENT
	)
}

func (transferOutput *TransferOutput) GetStorageKey() []byte {
	return transferOutput.id[:]
}

func (transferOutput *TransferOutput) Update(other objectstorage.StorableObject) {
	panic("TransferOutput should never be overwritten / updated")
}

func (transferOutput *TransferOutput) MarshalBinary() ([]byte, error) {
	transferOutput.realityIdMutex.RLock()

	balanceCount := len(transferOutput.balances)

	result := make([]byte, 4+balanceCount*coloredcoins.BalanceLength)
	offset := 0

	binary.LittleEndian.PutUint32(result[offset:], uint32(balanceCount))
	offset += 4

	for i := 0; i < balanceCount; i++ {
		if marshaledColoredBalance, err := transferOutput.balances[i].MarshalBinary(); err != nil {
			return nil, err
		} else {
			copy(result[offset:], marshaledColoredBalance)
			offset += coloredcoins.BalanceLength
		}
	}

	transferOutput.realityIdMutex.RUnlock()

	return result, nil
}

func (transferOutput *TransferOutput) UnmarshalBinary(serializedObject []byte) error {
	offset := 0

	if balances, err := transferOutput.unmarshalBalances(serializedObject, &offset); err != nil {
		return err
	} else {
		transferOutput.balances = balances
	}

	return nil
}

func (transferOutput *TransferOutput) unmarshalBalances(serializedBalances []byte, offset *int) ([]*coloredcoins.ColoredBalance, error) {
	balanceCount := int(binary.LittleEndian.Uint32(serializedBalances))
	*offset += 4

	balances := make([]*coloredcoins.ColoredBalance, balanceCount)
	for i := 0; i < balanceCount; i++ {
		coloredBalance := coloredcoins.ColoredBalance{}
		if err := coloredBalance.UnmarshalBinary(serializedBalances[4+i*coloredcoins.BalanceLength:]); err != nil {
			return nil, err
		}
		*offset += coloredcoins.BalanceLength

		balances[i] = &coloredBalance
	}

	return balances, nil
}

type CachedTransferOutput struct {
	objectstorage.CachedObject
}

func (cachedObject *CachedTransferOutput) Unwrap() *TransferOutput {
	if untypedObject := cachedObject.Get(); untypedObject == nil {
		return nil
	} else {
		if typedObject := untypedObject.(*TransferOutput); typedObject == nil || typedObject.IsDeleted() {
			return nil
		} else {
			return typedObject
		}
	}
}
