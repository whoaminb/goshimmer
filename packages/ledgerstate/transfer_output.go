package ledgerstate

import (
	"encoding/binary"
	"sync"

	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"

	"github.com/iotaledger/goshimmer/packages/binary/transfer"

	"github.com/iotaledger/goshimmer/packages/binary/address"

	"github.com/iotaledger/goshimmer/packages/stringify"
	"github.com/iotaledger/hive.go/objectstorage"
)

type TransferOutput struct {
	objectstorage.StorableObjectFlags

	transferHash transfer.Hash
	addressHash  address.Address
	balances     []*ColoredBalance
	realityId    reality.Id
	consumers    map[transfer.Hash][]address.Address

	storageKey  []byte
	ledgerState *LedgerState

	realityIdMutex sync.RWMutex
	consumersMutex sync.RWMutex
	bookingMutex   sync.Mutex
}

func NewTransferOutput(ledgerState *LedgerState, realityId reality.Id, transferHash transfer.Hash, addressHash address.Address, balances ...*ColoredBalance) *TransferOutput {
	return &TransferOutput{
		transferHash: transferHash,
		addressHash:  addressHash,
		balances:     balances,
		realityId:    realityId,
		consumers:    make(map[transfer.Hash][]address.Address),

		storageKey:  append(transferHash[:], addressHash[:]...),
		ledgerState: ledgerState,
	}
}

func (transferOutput *TransferOutput) GetTransferHash() (transferHash transfer.Hash) {
	transferHash = transferOutput.transferHash

	return
}

func (transferOutput *TransferOutput) GetRealityId() (realityId reality.Id) {
	transferOutput.realityIdMutex.RLock()
	realityId = transferOutput.realityId
	transferOutput.realityIdMutex.RUnlock()

	return
}

func (transferOutput *TransferOutput) GetAddressHash() (addressHash address.Address) {
	return transferOutput.addressHash
}

func (transferOutput *TransferOutput) SetRealityId(realityId reality.Id) {
	transferOutput.realityIdMutex.RLock()
	if transferOutput.realityId != realityId {
		transferOutput.realityIdMutex.RUnlock()

		transferOutput.realityIdMutex.Lock()
		if transferOutput.realityId != realityId {
			transferOutput.realityId = realityId

			transferOutput.SetModified()
		}
		transferOutput.realityIdMutex.Unlock()
	} else {
		transferOutput.realityIdMutex.RUnlock()
	}
}

func (transferOutput *TransferOutput) GetBalances() []*ColoredBalance {
	return transferOutput.balances
}

func (transferOutput *TransferOutput) GetConsumers() (consumers map[transfer.Hash][]address.Address) {
	consumers = make(map[transfer.Hash][]address.Address)

	transferOutput.consumersMutex.RLock()
	for transferHash, addresses := range transferOutput.consumers {
		consumers[transferHash] = make([]address.Address, len(addresses))
		copy(consumers[transferHash], addresses)
	}
	transferOutput.consumersMutex.RUnlock()

	return
}

func (transferOutput *TransferOutput) addConsumer(consumer transfer.Hash, outputs map[address.Address][]*ColoredBalance) (consumersToElevate map[transfer.Hash][]address.Address, err error) {
	transferOutput.consumersMutex.RLock()
	if _, exist := transferOutput.consumers[consumer]; exist {
		transferOutput.consumersMutex.RUnlock()
	} else {
		transferOutput.consumersMutex.RUnlock()

		transferOutput.consumersMutex.Lock()
		switch len(transferOutput.consumers) {
		case 0:
			consumersToElevate = nil
			err = transferOutput.markAsSpent()
		case 1:
			consumersToElevate = make(map[transfer.Hash][]address.Address, 1)
			for transferHash, addresses := range transferOutput.consumers {
				consumersToElevate[transferHash] = addresses
			}
			err = nil
		default:
			consumersToElevate = make(map[transfer.Hash][]address.Address)
			err = nil
		}
		consumers := make([]address.Address, len(outputs))
		i := 0
		for addressHash := range outputs {
			consumers[i] = addressHash

			i++
		}

		transferOutput.consumers[consumer] = consumers
		transferOutput.consumersMutex.Unlock()

		transferOutput.SetModified()
	}

	return
}

func (transferOutput *TransferOutput) markAsSpent() error {
	transferOutput.bookingMutex.Lock()

	currentBookingKey := generateTransferOutputBookingStorageKey(transferOutput.GetRealityId(), transferOutput.addressHash, false, transferOutput.transferHash)
	if oldTransferOutputBooking, err := transferOutput.ledgerState.transferOutputBookings.Load(currentBookingKey); err != nil {
		transferOutput.bookingMutex.Unlock()

		return err
	} else {
		transferOutput.ledgerState.storeTransferOutputBooking(newTransferOutputBooking(transferOutput.GetRealityId(), transferOutput.addressHash, true, transferOutput.transferHash)).Release()

		oldTransferOutputBooking.Consume(func(transferOutputBooking objectstorage.StorableObject) {
			transferOutputBooking.Delete()
		})
	}

	transferOutput.bookingMutex.Unlock()

	return nil
}

func (transferOutput *TransferOutput) String() string {
	return stringify.Struct("TransferOutput",
		stringify.StructField("transferHash", transferOutput.transferHash.String()),
		stringify.StructField("addressHash", transferOutput.addressHash.String()),
		stringify.StructField("balances", transferOutput.balances),
		stringify.StructField("realityId", transferOutput.realityId.String()),
		stringify.StructField("spent", len(transferOutput.consumers) >= 1),
	)
}

// region support object storage ///////////////////////////////////////////////////////////////////////////////////////

func (transferOutput *TransferOutput) GetStorageKey() []byte {
	return transferOutput.storageKey
}

func (transferOutput *TransferOutput) Update(other objectstorage.StorableObject) {}

func (transferOutput *TransferOutput) MarshalBinary() ([]byte, error) {
	transferOutput.realityIdMutex.RLock()
	transferOutput.consumersMutex.RLock()

	balanceCount := len(transferOutput.balances)
	consumerCount := len(transferOutput.consumers)

	serializedLength := reality.IdLength + 4 + balanceCount*coloredBalanceLength + 4 + consumerCount*transfer.HashLength
	for _, addresses := range transferOutput.consumers {
		serializedLength += 4
		for range addresses {
			serializedLength += address.Length
		}
	}

	result := make([]byte, serializedLength)

	copy(result[0:], transferOutput.realityId[:])

	binary.LittleEndian.PutUint32(result[reality.IdLength:], uint32(balanceCount))
	for i := 0; i < balanceCount; i++ {
		copy(result[reality.IdLength+4+i*coloredBalanceLength:], transferOutput.balances[i].color[:colorLength])
		binary.LittleEndian.PutUint64(result[reality.IdLength+4+i*coloredBalanceLength+colorLength:], transferOutput.balances[i].balance)
	}
	offset := reality.IdLength + 4 + balanceCount*coloredBalanceLength

	binary.LittleEndian.PutUint32(result[offset:], uint32(consumerCount))
	offset += 4
	for transferHash, addresses := range transferOutput.consumers {
		copy(result[offset:], transferHash[:transfer.HashLength])
		offset += transfer.HashLength

		binary.LittleEndian.PutUint32(result[offset:], uint32(len(addresses)))
		offset += 4

		for _, addressHash := range addresses {
			copy(result[offset:], addressHash[:address.Length])
			offset += address.Length
		}
	}

	transferOutput.consumersMutex.RUnlock()
	transferOutput.realityIdMutex.RUnlock()

	return result, nil
}

func (transferOutput *TransferOutput) UnmarshalBinary(serializedObject []byte) error {
	if err := transferOutput.transferHash.UnmarshalBinary(transferOutput.storageKey[:transfer.HashLength]); err != nil {
		return err
	}

	if err := transferOutput.addressHash.UnmarshalBinary(transferOutput.storageKey[transfer.HashLength:]); err != nil {
		return err
	}

	if err := transferOutput.realityId.UnmarshalBinary(serializedObject[:reality.IdLength]); err != nil {
		return err
	}

	if balances, err := transferOutput.unmarshalBalances(serializedObject[reality.IdLength:]); err != nil {
		return err
	} else {
		transferOutput.balances = balances
	}

	if consumers, err := transferOutput.unmarshalConsumers(serializedObject[reality.IdLength+4+len(transferOutput.balances)*coloredBalanceLength:]); err != nil {
		return err
	} else {
		transferOutput.consumers = consumers
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

func (transferOutput *TransferOutput) unmarshalConsumers(serializedConsumers []byte) (map[transfer.Hash][]address.Address, error) {
	offset := 0

	consumerCount := int(binary.LittleEndian.Uint32(serializedConsumers[offset:]))
	offset += 4

	consumers := make(map[transfer.Hash][]address.Address, consumerCount)
	for i := 0; i < consumerCount; i++ {
		transferHash := transfer.Hash{}
		if err := transferHash.UnmarshalBinary(serializedConsumers[offset:]); err != nil {
			return nil, err
		}
		offset += transfer.HashLength

		addressHashCount := int(binary.LittleEndian.Uint32(serializedConsumers[offset:]))
		offset += 4

		consumers[transferHash] = make([]address.Address, addressHashCount)
		for i := 0; i < addressHashCount; i++ {
			addressHash := address.Address{}
			if err := addressHash.UnmarshalBinary(serializedConsumers[offset:]); err != nil {
				return nil, err
			}
			offset += address.Length

			consumers[transferHash][i] = addressHash
		}
	}

	return consumers, nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
