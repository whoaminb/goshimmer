package mana

import (
	"sync"

	"github.com/iotaledger/goshimmer/packages/marshaling"

	"github.com/golang/protobuf/proto"
	"github.com/iotaledger/goshimmer/packages/errors"
	manaproto "github.com/iotaledger/goshimmer/packages/mana/proto"
)

type Transfer struct {
	inputs          []*Input
	inputsMutex     sync.RWMutex
	spentTime       uint64
	spentTimeMutex  sync.RWMutex
	burnedMana      uint64
	burnedManaMutex sync.RWMutex
}

func NewTransfer(inputs []*Input, spentTime uint64, burnedMana uint64) *Transfer {
	return &Transfer{
		inputs:     inputs,
		spentTime:  spentTime,
		burnedMana: burnedMana,
	}
}

// Returns a protobuf representation of this transfer.
func (transfer *Transfer) ToProto() (result proto.Message) {
	transfer.inputsMutex.RLock()
	transfer.spentTimeMutex.RLock()
	transfer.burnedManaMutex.RLock()
	defer transfer.inputsMutex.RUnlock()
	defer transfer.spentTimeMutex.RUnlock()
	defer transfer.burnedManaMutex.RUnlock()

	protoTransfer := &manaproto.Transfer{
		Inputs:     make([]*manaproto.Input, len(transfer.inputs)),
		SpentTime:  transfer.spentTime,
		BurnedMana: transfer.burnedMana,
	}

	for i, input := range transfer.inputs {
		protoTransfer.Inputs[i] = input.ToProto()
	}

	return protoTransfer
}

// Restores the values from a protobuf representation of a transfer.
func (transfer *Transfer) FromProto(proto proto.Message) {
	transfer.inputsMutex.Lock()
	transfer.spentTimeMutex.Lock()
	transfer.burnedManaMutex.Lock()
	defer transfer.inputsMutex.Unlock()
	defer transfer.spentTimeMutex.Unlock()
	defer transfer.burnedManaMutex.Unlock()

	protoTransfer := proto.(*manaproto.Transfer)

	transfer.inputs = make([]*Input, len(protoTransfer.Inputs))
	transfer.spentTime = protoTransfer.SpentTime
	transfer.burnedMana = protoTransfer.BurnedMana

	for i, protoInput := range protoTransfer.Inputs {
		var input Input
		input.FromProto(protoInput)

		transfer.inputs[i] = &input
	}
}

func (transfer *Transfer) MarshalBinary() ([]byte, errors.IdentifiableError) {
	return marshaling.Marshal(transfer)
}

func (transfer *Transfer) UnmarshalBinary(data []byte) (err errors.IdentifiableError) {
	return marshaling.Unmarshal(transfer, &manaproto.Transfer{}, data)
}
