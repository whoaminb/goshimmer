package mana

import (
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/iotaledger/goshimmer/packages/marshaling"

	"github.com/iotaledger/goshimmer/packages/errors"
	manaproto "github.com/iotaledger/goshimmer/packages/mana/proto"
)

type Input struct {
	coinAmount        uint64
	coinAmountMutex   sync.RWMutex
	receivedTime      uint64
	receivedTimeMutex sync.RWMutex
}

func NewInput(coinAmount uint64, receivedTime uint64) *Input {
	return &Input{
		coinAmount:   coinAmount,
		receivedTime: receivedTime,
	}
}

func (input *Input) GetCoinAmount() uint64 {
	input.coinAmountMutex.RLock()
	defer input.coinAmountMutex.RUnlock()

	return input.coinAmount
}

func (input *Input) SetCoinAmount(coinAmount uint64) {
	input.coinAmountMutex.Lock()
	defer input.coinAmountMutex.Unlock()

	input.coinAmount = coinAmount
}

func (input *Input) GetReceivedTime() uint64 {
	input.receivedTimeMutex.RLock()
	defer input.receivedTimeMutex.RUnlock()

	return input.receivedTime
}

func (input *Input) SetReceivedTime(receivedTime uint64) {
	input.receivedTimeMutex.Lock()
	defer input.receivedTimeMutex.Unlock()

	input.receivedTime = receivedTime
}

func (input *Input) ToProto() (result proto.Message) {
	input.receivedTimeMutex.RLock()
	input.coinAmountMutex.RLock()
	defer input.receivedTimeMutex.RUnlock()
	defer input.coinAmountMutex.RUnlock()

	return &manaproto.Input{
		CoinAmount:   input.coinAmount,
		ReceivedTime: input.receivedTime,
	}
}

func (input *Input) FromProto(proto proto.Message) {
	input.receivedTimeMutex.Lock()
	input.coinAmountMutex.Lock()
	defer input.receivedTimeMutex.Unlock()
	defer input.coinAmountMutex.Unlock()

	inputProto := proto.(*manaproto.Input)

	input.coinAmount = inputProto.CoinAmount
	input.receivedTime = inputProto.ReceivedTime
}

func (input *Input) MarshalBinary() ([]byte, errors.IdentifiableError) {
	return marshaling.Marshal(input)
}

func (input *Input) UnmarshalBinary(data []byte) errors.IdentifiableError {
	return marshaling.Unmarshal(input, data, &manaproto.Input{})
}

func (input *Input) Equals(other *Input) bool {
	if input == other {
		return true
	}

	if input == nil || other == nil {
		return false
	}

	input.receivedTimeMutex.RLock()
	input.coinAmountMutex.RLock()
	other.receivedTimeMutex.RLock()
	other.coinAmountMutex.RLock()
	defer input.receivedTimeMutex.RUnlock()
	defer input.coinAmountMutex.RUnlock()
	defer other.receivedTimeMutex.RUnlock()
	defer other.coinAmountMutex.RUnlock()

	return input.coinAmount == other.coinAmount && input.receivedTime == other.receivedTime
}
