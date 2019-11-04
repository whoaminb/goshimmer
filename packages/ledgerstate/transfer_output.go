package ledgerstate

import (
	"encoding/binary"
	"sync"

	"github.com/iotaledger/goshimmer/packages/objectstorage"
	"github.com/iotaledger/goshimmer/packages/stringify"
)

type TransferOutput struct {
	transferHash TransferHash
	addressHash  AddressHash
	balances     []*ColoredBalance
	realityId    RealityId
	consumers    map[TransferHash]empty

	storageKey  []byte
	ledgerState *LedgerState

	realityIdMutex sync.RWMutex
	consumersMutex sync.RWMutex
}

func NewTransferOutput(ledgerState *LedgerState, realityId RealityId, transferHash TransferHash, addressHash AddressHash, balances ...*ColoredBalance) *TransferOutput {
	return &TransferOutput{
		transferHash: transferHash,
		addressHash:  addressHash,
		balances:     balances,
		realityId:    realityId,
		consumers:    make(map[TransferHash]empty),

		storageKey:  append(transferHash[:], addressHash[:]...),
		ledgerState: ledgerState,
	}
}

func (transferOutput *TransferOutput) GetRealityId() (realityId RealityId) {
	transferOutput.realityIdMutex.RLock()
	realityId = transferOutput.realityId
	transferOutput.realityIdMutex.RUnlock()

	return
}

func (transferOutput *TransferOutput) SetRealityId(realityId RealityId) {
	transferOutput.realityIdMutex.Lock()
	transferOutput.realityId = realityId
	transferOutput.realityIdMutex.Unlock()
}

func (transferOutput *TransferOutput) GetBalances() []*ColoredBalance {
	return transferOutput.balances
}

func (transferOutput *TransferOutput) addConsumer(consumer TransferHash) bool {
	transferOutput.consumersMutex.RLock()
	if _, exist := transferOutput.consumers[consumer]; exist {
		transferOutput.consumersMutex.RUnlock()

		return false
	} else {
		transferOutput.consumersMutex.RUnlock()

		transferOutput.consumersMutex.Lock()
		if len(transferOutput.consumers) == 0 {
			transferOutput.markAsSpent()
		} else {
			panic("DOUBLE SPEND DETECTED")
		}
		transferOutput.consumers[consumer] = void
		transferOutput.consumersMutex.Unlock()

		return true
	}
}

func (transferOutput *TransferOutput) getConsumers() (consumers map[TransferHash]empty) {
	transferOutput.consumersMutex.RLock()
	consumers = transferOutput.consumers
	transferOutput.consumersMutex.RUnlock()

	return
}

func (transferOutput *TransferOutput) markAsSpent() {
	currentBookingKey := generateTransferOutputBookingStorageKey(transferOutput.realityId, transferOutput.addressHash, false, transferOutput.transferHash)

	if cachedTransferOutputBooking, err := transferOutput.ledgerState.transferOutputBookings.Load(currentBookingKey); err != nil {
		panic(err)
	} else if !cachedTransferOutputBooking.Exists() {
		panic("could not find TransferOutputBooking")
	} else {
		transferOutputBooking := cachedTransferOutputBooking.Get().(*TransferOutputBooking)
		transferOutput.ledgerState.storeTransferOutputBooking(newTransferOutputBooking(transferOutputBooking.GetRealityId(), transferOutputBooking.GetAddressHash(), true, transferOutputBooking.GetTransferHash())).Release()

		cachedTransferOutputBooking.Delete()
		cachedTransferOutputBooking.Release()
	}
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
	balanceCount := len(transferOutput.balances)
	consumerCount := len(transferOutput.consumers)

	result := make([]byte, realityIdLength+4+balanceCount*coloredBalanceLength+4+consumerCount*transferHashLength)

	copy(result[0:], transferOutput.realityId[:])

	binary.LittleEndian.PutUint32(result[realityIdLength:], uint32(balanceCount))
	for i := 0; i < balanceCount; i++ {
		copy(result[realityIdLength+4+i*coloredBalanceLength:], transferOutput.balances[i].color[:colorLength])
		binary.LittleEndian.PutUint64(result[realityIdLength+4+i*coloredBalanceLength+colorLength:], transferOutput.balances[i].balance)
	}

	offset := realityIdLength + 4 + balanceCount*coloredBalanceLength
	binary.LittleEndian.PutUint32(result[offset:], uint32(consumerCount))
	offset += 4
	for consumer := range transferOutput.consumers {
		copy(result[offset:], consumer[:transferHashLength])
		offset += transferHashLength
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

func (transferOutput *TransferOutput) unmarshalConsumers(serializedConsumers []byte) (map[TransferHash]empty, error) {
	consumerCount := int(binary.LittleEndian.Uint32(serializedConsumers))

	consumers := make(map[TransferHash]empty, consumerCount)
	for i := 0; i < consumerCount; i++ {
		transferHash := TransferHash{}
		if err := transferHash.UnmarshalBinary(serializedConsumers[4+i*transferHashLength:]); err != nil {
			return nil, err
		}

		consumers[transferHash] = void
	}

	return consumers, nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
