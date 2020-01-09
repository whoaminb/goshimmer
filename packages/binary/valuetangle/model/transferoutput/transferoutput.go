package transferoutput

import (
	"encoding/binary"
	"sync"

	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/types"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/coloredcoins"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"
	"github.com/iotaledger/goshimmer/packages/stringify"
	"github.com/iotaledger/hive.go/objectstorage"
)

type TransferOutput struct {
	objectstorage.StorableObjectFlags

	transactionId transaction.Id
	address       address.Address
	realityId     reality.Id
	balances      []*coloredcoins.ColoredBalance
	consumers     map[transaction.Id]types.Empty

	realityIdMutex sync.RWMutex
	consumersMutex sync.RWMutex
}

func NewTransferOutput(transactionId transaction.Id, address address.Address, balances ...*coloredcoins.ColoredBalance) *TransferOutput {
	return &TransferOutput{
		transactionId: transactionId,
		address:       address,
		balances:      balances,
		realityId:     reality.EmptyId,
		consumers:     make(map[transaction.Id]types.Empty),
	}
}

func FromStorage(key []byte) objectstorage.StorableObject {
	result := &TransferOutput{}
	offset := 0

	if err := result.transactionId.UnmarshalBinary(key[offset:]); err != nil {
		panic(err)
	}
	offset += transaction.IdLength

	if err := result.address.UnmarshalBinary(key[offset:]); err != nil {
		panic(err)
	}

	return result
}

func (transferOutput *TransferOutput) GetTransactionId() (transactionId transaction.Id) {
	transactionId = transferOutput.transactionId

	return
}

func (transferOutput *TransferOutput) GetAddress() (address address.Address) {
	return transferOutput.address
}

func (transferOutput *TransferOutput) GetRealityId() (realityId reality.Id) {
	transferOutput.realityIdMutex.RLock()
	realityId = transferOutput.realityId
	transferOutput.realityIdMutex.RUnlock()

	return
}

func (transferOutput *TransferOutput) SetRealityId(realityId reality.Id) (modified bool) {
	transferOutput.realityIdMutex.RLock()
	if transferOutput.realityId != realityId {
		transferOutput.realityIdMutex.RUnlock()

		transferOutput.realityIdMutex.Lock()
		if transferOutput.realityId != realityId {
			transferOutput.realityId = realityId

			transferOutput.SetModified()

			modified = true
		}
		transferOutput.realityIdMutex.Unlock()
	} else {
		transferOutput.realityIdMutex.RUnlock()
	}

	return
}

func (transferOutput *TransferOutput) GetBalances() []*coloredcoins.ColoredBalance {
	return transferOutput.balances
}

func (transferOutput *TransferOutput) GetConsumers() (consumers map[transaction.Id]types.Empty) {
	consumers = make(map[transaction.Id]types.Empty)

	transferOutput.consumersMutex.RLock()
	for transferHash := range transferOutput.consumers {
		consumers[transferHash] = types.Void
	}
	transferOutput.consumersMutex.RUnlock()

	return
}

func (transferOutput *TransferOutput) IsSpent() (result bool) {
	transferOutput.consumersMutex.RLock()
	result = len(transferOutput.consumers) >= 1
	transferOutput.consumersMutex.RUnlock()

	return
}

func (transferOutput *TransferOutput) AddConsumer(consumer transaction.Id) (previousConsumers map[transaction.Id]types.Empty) {
	transferOutput.consumersMutex.RLock()
	if _, exist := transferOutput.consumers[consumer]; !exist {
		transferOutput.consumersMutex.RUnlock()

		transferOutput.consumersMutex.Lock()
		if _, exist := transferOutput.consumers[consumer]; !exist {
			previousConsumers = make(map[transaction.Id]types.Empty)
			for transactionId := range transferOutput.consumers {
				previousConsumers[transactionId] = types.Void
			}

			transferOutput.consumers[consumer] = types.Void

			transferOutput.SetModified()
		}
		transferOutput.consumersMutex.Unlock()
	} else {
		transferOutput.consumersMutex.RUnlock()
	}

	return
}

func (transferOutput *TransferOutput) String() string {
	return stringify.Struct("TransferOutput",
		stringify.StructField("transactionId", transferOutput.GetTransactionId().String()),
		stringify.StructField("address", transferOutput.GetAddress().String()),
		stringify.StructField("realityId", transferOutput.GetRealityId().String()),
		stringify.StructField("balances", transferOutput.GetBalances()),
		stringify.StructField("spent", len(transferOutput.GetConsumers()) >= 1),
	)
}

func (transferOutput *TransferOutput) GetStorageKey() []byte {
	return append(transferOutput.transactionId[:], transferOutput.address[:]...)
}

func (transferOutput *TransferOutput) Update(other objectstorage.StorableObject) {
	panic("TransferOutput should never be overwritten / updated")
}

func (transferOutput *TransferOutput) MarshalBinary() ([]byte, error) {
	transferOutput.realityIdMutex.RLock()
	transferOutput.consumersMutex.RLock()

	balanceCount := len(transferOutput.balances)
	consumerCount := len(transferOutput.consumers)

	result := make([]byte, reality.IdLength+4+balanceCount*coloredcoins.BalanceLength+4+consumerCount*transaction.IdLength)
	offset := 0

	copy(result[0:], transferOutput.realityId[:])
	offset += reality.IdLength

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

	binary.LittleEndian.PutUint32(result[offset:], uint32(consumerCount))
	offset += 4

	for transactionId := range transferOutput.consumers {
		copy(result[offset:], transactionId[:])
		offset += transaction.IdLength
	}

	transferOutput.consumersMutex.RUnlock()
	transferOutput.realityIdMutex.RUnlock()

	return result, nil
}

func (transferOutput *TransferOutput) UnmarshalBinary(serializedObject []byte) error {
	offset := 0

	if err := transferOutput.realityId.UnmarshalBinary(serializedObject[offset:]); err != nil {
		return err
	}
	offset += reality.IdLength

	if balances, err := transferOutput.unmarshalBalances(serializedObject, &offset); err != nil {
		return err
	} else {
		transferOutput.balances = balances
	}

	if consumers, err := transferOutput.unmarshalConsumers(serializedObject, &offset); err != nil {
		return err
	} else {
		transferOutput.consumers = consumers
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

func (transferOutput *TransferOutput) unmarshalConsumers(serializedConsumers []byte, offset *int) (map[transaction.Id]types.Empty, error) {
	consumerCount := int(binary.LittleEndian.Uint32(serializedConsumers[*offset:]))
	*offset += 4

	consumers := make(map[transaction.Id]types.Empty, consumerCount)
	for i := 0; i < consumerCount; i++ {
		var transactionId transaction.Id
		if err := transactionId.UnmarshalBinary(serializedConsumers[*offset:]); err != nil {
			return nil, err
		}
		*offset += transaction.IdLength
	}

	return consumers, nil
}

type CachedTransferOutput struct {
	*objectstorage.CachedObject
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
