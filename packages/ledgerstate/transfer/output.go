package transfer

import (
	"encoding/binary"
	"sync"

	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/coloredcoins"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"
	"github.com/iotaledger/goshimmer/packages/stringify"

	"github.com/iotaledger/hive.go/objectstorage"
)

type Output struct {
	objectstorage.StorableObjectFlags

	transferHash Hash
	addressHash  address.Address
	balances     []*coloredcoins.ColoredBalance
	realityId    reality.Id
	consumers    map[Hash][]address.Address

	storageKey     []byte
	OutputBookings *objectstorage.ObjectStorage

	realityIdMutex sync.RWMutex
	consumersMutex sync.RWMutex
	bookingMutex   sync.Mutex
}

func NewTransferOutput(outputBookings *objectstorage.ObjectStorage, realityId reality.Id, transferHash Hash, addressHash address.Address, balances ...*coloredcoins.ColoredBalance) *Output {
	return &Output{
		transferHash: transferHash,
		addressHash:  addressHash,
		balances:     balances,
		realityId:    realityId,
		consumers:    make(map[Hash][]address.Address),

		storageKey:     append(transferHash[:], addressHash[:]...),
		OutputBookings: outputBookings,
	}
}

func (transferOutput *Output) GetTransferHash() (transferHash Hash) {
	transferHash = transferOutput.transferHash

	return
}

func (transferOutput *Output) GetRealityId() (realityId reality.Id) {
	transferOutput.realityIdMutex.RLock()
	realityId = transferOutput.realityId
	transferOutput.realityIdMutex.RUnlock()

	return
}

func (transferOutput *Output) GetAddressHash() (addressHash address.Address) {
	return transferOutput.addressHash
}

func (transferOutput *Output) SetRealityId(realityId reality.Id) {
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

func (transferOutput *Output) GetBalances() []*coloredcoins.ColoredBalance {
	return transferOutput.balances
}

func (transferOutput *Output) GetConsumers() (consumers map[Hash][]address.Address) {
	consumers = make(map[Hash][]address.Address)

	transferOutput.consumersMutex.RLock()
	for transferHash, addresses := range transferOutput.consumers {
		consumers[transferHash] = make([]address.Address, len(addresses))
		copy(consumers[transferHash], addresses)
	}
	transferOutput.consumersMutex.RUnlock()

	return
}

func (transferOutput *Output) AddConsumer(consumer Hash, outputs map[address.Address][]*coloredcoins.ColoredBalance) (consumersToElevate map[Hash][]address.Address, err error) {
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
			consumersToElevate = make(map[Hash][]address.Address, 1)
			for transferHash, addresses := range transferOutput.consumers {
				consumersToElevate[transferHash] = addresses
			}
			err = nil
		default:
			consumersToElevate = make(map[Hash][]address.Address)
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

func (transferOutput *Output) markAsSpent() error {
	transferOutput.bookingMutex.Lock()

	currentBookingKey := GenerateOutputBookingStorageKey(transferOutput.GetRealityId(), transferOutput.addressHash, false, transferOutput.transferHash)
	oldTransferOutputBooking := transferOutput.OutputBookings.Load(currentBookingKey)
	transferOutput.OutputBookings.Store(NewTransferOutputBooking(transferOutput.GetRealityId(), transferOutput.addressHash, true, transferOutput.transferHash)).Release()

	oldTransferOutputBooking.Consume(func(transferOutputBooking objectstorage.StorableObject) {
		transferOutputBooking.Delete()
	})

	transferOutput.bookingMutex.Unlock()

	return nil
}

func (transferOutput *Output) String() string {
	return stringify.Struct("Output",
		stringify.StructField("transferHash", transferOutput.transferHash.String()),
		stringify.StructField("addressHash", transferOutput.addressHash.String()),
		stringify.StructField("balances", transferOutput.balances),
		stringify.StructField("realityId", transferOutput.realityId.String()),
		stringify.StructField("spent", len(transferOutput.consumers) >= 1),
	)
}

// region support object storage ///////////////////////////////////////////////////////////////////////////////////////

func (transferOutput *Output) GetStorageKey() []byte {
	return transferOutput.storageKey
}

func (transferOutput *Output) Update(other objectstorage.StorableObject) {}

func (transferOutput *Output) MarshalBinary() ([]byte, error) {
	transferOutput.realityIdMutex.RLock()
	transferOutput.consumersMutex.RLock()

	balanceCount := len(transferOutput.balances)
	consumerCount := len(transferOutput.consumers)

	serializedLength := reality.IdLength + 4 + balanceCount*coloredcoins.BalanceLength + 4 + consumerCount*HashLength
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
		color := transferOutput.balances[i].GetColor()

		copy(result[reality.IdLength+4+i*coloredcoins.BalanceLength:], color[:coloredcoins.ColorLength])
		binary.LittleEndian.PutUint64(result[reality.IdLength+4+i*coloredcoins.BalanceLength+coloredcoins.ColorLength:], transferOutput.balances[i].GetBalance())
	}
	offset := reality.IdLength + 4 + balanceCount*coloredcoins.BalanceLength

	binary.LittleEndian.PutUint32(result[offset:], uint32(consumerCount))
	offset += 4
	for transferHash, addresses := range transferOutput.consumers {
		copy(result[offset:], transferHash[:HashLength])
		offset += HashLength

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

func (transferOutput *Output) UnmarshalBinary(serializedObject []byte) error {
	if err := transferOutput.transferHash.UnmarshalBinary(transferOutput.storageKey[:HashLength]); err != nil {
		return err
	}

	if err := transferOutput.addressHash.UnmarshalBinary(transferOutput.storageKey[HashLength:]); err != nil {
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

	if consumers, err := transferOutput.unmarshalConsumers(serializedObject[reality.IdLength+4+len(transferOutput.balances)*coloredcoins.BalanceLength:]); err != nil {
		return err
	} else {
		transferOutput.consumers = consumers
	}

	return nil
}

func (transferOutput *Output) IsSpent() (result bool) {
	transferOutput.consumersMutex.RLock()
	result = len(transferOutput.consumers) >= 1
	transferOutput.consumersMutex.RUnlock()

	return
}

func (transferOutput *Output) unmarshalBalances(serializedBalances []byte) ([]*coloredcoins.ColoredBalance, error) {
	balanceCount := int(binary.LittleEndian.Uint32(serializedBalances))

	balances := make([]*coloredcoins.ColoredBalance, balanceCount)
	for i := 0; i < balanceCount; i++ {
		coloredBalance := coloredcoins.ColoredBalance{}
		if err := coloredBalance.UnmarshalBinary(serializedBalances[4+i*coloredcoins.BalanceLength:]); err != nil {
			return nil, err
		}

		balances[i] = &coloredBalance
	}

	return balances, nil
}

func (transferOutput *Output) unmarshalConsumers(serializedConsumers []byte) (map[Hash][]address.Address, error) {
	offset := 0

	consumerCount := int(binary.LittleEndian.Uint32(serializedConsumers[offset:]))
	offset += 4

	consumers := make(map[Hash][]address.Address, consumerCount)
	for i := 0; i < consumerCount; i++ {
		transferHash := Hash{}
		if err := transferHash.UnmarshalBinary(serializedConsumers[offset:]); err != nil {
			return nil, err
		}
		offset += HashLength

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

func OutputFactory(key []byte) objectstorage.StorableObject {
	result := &Output{
		storageKey: make([]byte, len(key)),
	}
	copy(result.storageKey, key)

	return result
}
