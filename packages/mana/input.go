package mana

import (
	"encoding/binary"
	"fmt"

	"github.com/iotaledger/goshimmer/packages/marshal"

	"github.com/iotaledger/goshimmer/packages/errors"
)

type Input struct {
	coinAmount   uint64
	receivedTime uint64
}

func (input *Input) MarshalBinary() (data []byte, err errors.IdentifiableError) {
	data = inputSchema.Marshal(input)

	return
}

func (input *Input) UnmarshalBinary(data []byte) (err errors.IdentifiableError) {
	if len(data) < INPUT_TOTAL_MARSHALED_SIZE {
		err = ErrUnmarshalFailed.Derive("byte sequence of marshaled input is not long enough")
	}

	input.coinAmount = binary.BigEndian.Uint64(data[INPUT_COIN_AMOUNT_BYTE_OFFSET_START:INPUT_COIN_AMOUNT_BYTE_OFFSET_END])
	input.receivedTime = binary.BigEndian.Uint64(data[INPUT_RECEIVED_TIME_BYTE_OFFSET_START:INPUT_RECEIVED_TIME_BYTE_OFFSET_END])

	return
}

func (input *Input) GetCoinAmount() uint64 {
	return input_getCoinAmount(input)
}

func input_getCoinAmount(input interface{}) uint64 {
	return input.(*Input).coinAmount
}

func input_setCoinAmount(input interface{}, coinAmount uint64) {
	input.(*Input).coinAmount = coinAmount
}

func input_getReceivedTime(input interface{}) uint64 {
	return input.(*Input).receivedTime
}

func input_setReceivedTime(input interface{}, receivedTime uint64) {
	input.(*Input).coinAmount = receivedTime
}

var inputSchema = marshal.Schema(
	marshal.Uint64(input_getCoinAmount, input_setCoinAmount),
	marshal.Uint64(input_getReceivedTime, input_setReceivedTime),
)

func init() {
	fmt.Println((&Input{
		coinAmount:   10,
		receivedTime: 20,
	}).MarshalBinary())
}

const (
	INPUT_COIN_AMOUNT_BYTE_OFFSET_START  = 0
	INPUT_COIN_AMOUNT_BYTE_OFFSET_LENGTH = 8
	INPUT_COIN_AMOUNT_BYTE_OFFSET_END    = INPUT_COIN_AMOUNT_BYTE_OFFSET_START + INPUT_COIN_AMOUNT_BYTE_OFFSET_LENGTH

	INPUT_RECEIVED_TIME_BYTE_OFFSET_START  = INPUT_COIN_AMOUNT_BYTE_OFFSET_END
	INPUT_RECEIVED_TIME_BYTE_OFFSET_LENGTH = 8
	INPUT_RECEIVED_TIME_BYTE_OFFSET_END    = INPUT_RECEIVED_TIME_BYTE_OFFSET_START + INPUT_RECEIVED_TIME_BYTE_OFFSET_LENGTH

	INPUT_TOTAL_MARSHALED_SIZE = INPUT_RECEIVED_TIME_BYTE_OFFSET_END
)
