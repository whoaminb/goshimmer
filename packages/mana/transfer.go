package mana

import (
	"github.com/iotaledger/goshimmer/packages/errors"
)

type Transfer struct {
	inputs     []*Input
	spentTime  uint64
	burnedMana uint64
}

func (transfer *Transfer) MarshalBinary() (data []byte, err errors.IdentifiableError) {
	data = make([]byte, INPUT_TOTAL_MARSHALED_SIZE)

	return
}

func (transfer *Transfer) UnmarshalBinary(data []byte) (err errors.IdentifiableError) {
	return
}

const (
	TRANSFER_SPENT_TIME_BYTE_OFFSET_START  = 0
	TRANSFER_SPENT_TIME_BYTE_OFFSET_LENGTH = 8
	TRANSFER_SPENT_TIME_BYTE_OFFSET_END    = TRANSFER_SPENT_TIME_BYTE_OFFSET_START + TRANSFER_SPENT_TIME_BYTE_OFFSET_LENGTH

	TRANSFER_BURNED_MANA_BYTE_OFFSET_START  = INPUT_COIN_AMOUNT_BYTE_OFFSET_END
	TRANSFER_BURNED_MANA_BYTE_OFFSET_LENGTH = 8
	TRANSFER_BURNED_MANA_BYTE_OFFSET_END    = INPUT_RECEIVED_TIME_BYTE_OFFSET_START + INPUT_RECEIVED_TIME_BYTE_OFFSET_LENGTH
)
