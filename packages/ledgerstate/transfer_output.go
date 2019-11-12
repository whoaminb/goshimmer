package ledgerstate

import (
	"encoding/binary"
	"sync"

	"github.com/iotaledger/goshimmer/packages/stringify"
	"github.com/iotaledger/hive.go/objectstorage"
)

type TransferOutput struct {
	transferHash TransferHash
	addressHash  AddressHash
	balances     []*ColoredBalance
	realityId    RealityId
	consumers    map[TransferHash][]AddressHash

	storageKey  []byte
	ledgerState *LedgerState

	realityIdMutex sync.RWMutex
	consumersMutex sync.RWMutex
	bookingMutex   sync.Mutex
}

func NewTransferOutput(ledgerState *LedgerState, realityId RealityId, transferHash TransferHash, addressHash AddressHash, balances ...*ColoredBalance) *TransferOutput {
	return &TransferOutput{
		transferHash: transferHash,
		addressHash:  addressHash,
		balances:     balances,
		realityId:    realityId,
		consumers:    make(map[TransferHash][]AddressHash),

		storageKey:  append(transferHash[:], addressHash[:]...),
		ledgerState: ledgerState,
	}
}

func (transferOutput *TransferOutput) GetTransferHash() (transferHash TransferHash) {
	transferHash = transferOutput.transferHash

	return
}

func (transferOutput *TransferOutput) GetRealityId() (realityId RealityId) {
	transferOutput.realityIdMutex.RLock()
	realityId = transferOutput.realityId
	transferOutput.realityIdMutex.RUnlock()

	return
}

func (transferOutput *TransferOutput) GetAddressHash() (addressHash AddressHash) {
	return transferOutput.addressHash
}

func (transferOutput *TransferOutput) SetRealityId(realityId RealityId) {
	transferOutput.realityIdMutex.Lock()
	transferOutput.realityId = realityId
	transferOutput.realityIdMutex.Unlock()
}

func (transferOutput *TransferOutput) GetBalances() []*ColoredBalance {
	return transferOutput.balances
}

func (transferOutput *TransferOutput) GetConsumers() (consumers map[TransferHash][]AddressHash) {
	consumers = make(map[TransferHash][]AddressHash)

	transferOutput.consumersMutex.RLock()
	for transferHash, addresses := range transferOutput.consumers {
		consumers[transferHash] = make([]AddressHash, len(addresses))
		copy(consumers[transferHash], addresses)
	}
	transferOutput.consumersMutex.RUnlock()

	return
}

func (transferOutput *TransferOutput) addConsumer(consumer TransferHash, outputs map[AddressHash][]*ColoredBalance) (isConflicting bool, consumersToElevate map[TransferHash][]AddressHash, err error) {
	transferOutput.consumersMutex.RLock()
	if _, exist := transferOutput.consumers[consumer]; exist {
		transferOutput.consumersMutex.RUnlock()
	} else {
		transferOutput.consumersMutex.RUnlock()

		transferOutput.consumersMutex.Lock()
		switch len(transferOutput.consumers) {
		case 0:
			isConflicting = false
			consumersToElevate = nil
			err = transferOutput.markAsSpent()
		case 1:
			isConflicting = true
			consumersToElevate = make(map[TransferHash][]AddressHash, 1)
			for transferHash, addresses := range transferOutput.consumers {
				consumersToElevate[transferHash] = addresses
			}
			err = nil
		default:
			isConflicting = true
			consumersToElevate = nil
			err = nil
		}
		consumers := make([]AddressHash, len(outputs))
		i := 0
		for addressHash := range outputs {
			consumers[i] = addressHash

			i++
		}

		transferOutput.consumers[consumer] = consumers
		transferOutput.consumersMutex.Unlock()
	}

	return
}

func (transferOutput *TransferOutput) moveToReality(realityId RealityId) error {
	transferOutput.bookingMutex.Lock()

	currentBookingKey := generateTransferOutputBookingStorageKey(transferOutput.GetRealityId(), transferOutput.addressHash, len(transferOutput.consumers) >= 1, transferOutput.transferHash)
	if oldTransferOutputBooking, err := transferOutput.ledgerState.transferOutputBookings.Load(currentBookingKey); err != nil {
		transferOutput.bookingMutex.Unlock()

		return err
	} else {
		transferOutput.realityIdMutex.Lock()
		transferOutput.realityId = realityId
		transferOutput.realityIdMutex.Unlock()

		transferOutput.ledgerState.storeTransferOutputBooking(newTransferOutputBooking(realityId, transferOutput.addressHash, len(transferOutput.consumers) >= 1, transferOutput.transferHash)).Release()

		oldTransferOutputBooking.Delete().Release()
	}

	transferOutput.bookingMutex.Unlock()

	return nil
}

func (transferOutput *TransferOutput) markAsSpent() error {
	transferOutput.bookingMutex.Lock()

	currentBookingKey := generateTransferOutputBookingStorageKey(transferOutput.GetRealityId(), transferOutput.addressHash, false, transferOutput.transferHash)
	if oldTransferOutputBooking, err := transferOutput.ledgerState.transferOutputBookings.Load(currentBookingKey); err != nil {
		transferOutput.bookingMutex.Unlock()

		return err
	} else {
		transferOutput.ledgerState.storeTransferOutputBooking(newTransferOutputBooking(transferOutput.GetRealityId(), transferOutput.addressHash, true, transferOutput.transferHash)).Release()

		oldTransferOutputBooking.Delete().Release()
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

	serializedLength := realityIdLength + 4 + balanceCount*coloredBalanceLength + 4 + consumerCount*transferHashLength
	for _, addresses := range transferOutput.consumers {
		serializedLength += 4
		for range addresses {
			serializedLength += addressHashLength
		}
	}

	result := make([]byte, serializedLength)

	copy(result[0:], transferOutput.realityId[:])

	binary.LittleEndian.PutUint32(result[realityIdLength:], uint32(balanceCount))
	for i := 0; i < balanceCount; i++ {
		copy(result[realityIdLength+4+i*coloredBalanceLength:], transferOutput.balances[i].color[:colorLength])
		binary.LittleEndian.PutUint64(result[realityIdLength+4+i*coloredBalanceLength+colorLength:], transferOutput.balances[i].balance)
	}
	offset := realityIdLength + 4 + balanceCount*coloredBalanceLength

	binary.LittleEndian.PutUint32(result[offset:], uint32(consumerCount))
	offset += 4
	for transferHash, addresses := range transferOutput.consumers {
		copy(result[offset:], transferHash[:transferHashLength])
		offset += transferHashLength

		binary.LittleEndian.PutUint32(result[offset:], uint32(len(addresses)))
		offset += 4

		for _, addressHash := range addresses {
			copy(result[offset:], addressHash[:addressHashLength])
			offset += addressHashLength

		}
	}

	transferOutput.consumersMutex.RUnlock()
	transferOutput.realityIdMutex.RUnlock()

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

	if consumers, err := transferOutput.unmarshalConsumers(serializedObject[realityIdLength+4+len(transferOutput.balances)*coloredBalanceLength:]); err != nil {
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

func (transferOutput *TransferOutput) unmarshalConsumers(serializedConsumers []byte) (map[TransferHash][]AddressHash, error) {
	offset := 0

	consumerCount := int(binary.LittleEndian.Uint32(serializedConsumers[offset:]))
	offset += 4

	consumers := make(map[TransferHash][]AddressHash, consumerCount)
	for i := 0; i < consumerCount; i++ {
		transferHash := TransferHash{}
		if err := transferHash.UnmarshalBinary(serializedConsumers[offset:]); err != nil {
			return nil, err
		}
		offset += transferHashLength

		addressHashCount := int(binary.LittleEndian.Uint32(serializedConsumers[offset:]))
		offset += 4

		consumers[transferHash] = make([]AddressHash, addressHashCount)
		for i := 0; i < addressHashCount; i++ {
			addressHash := AddressHash{}
			if err := addressHash.UnmarshalBinary(serializedConsumers[offset:]); err != nil {
				return nil, err
			}
			offset += addressHashLength

			consumers[transferHash][i] = addressHash
		}
	}

	return consumers, nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
