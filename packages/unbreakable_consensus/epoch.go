package unbreakable_consensus

import (
	"github.com/iotaledger/goshimmer/packages/stringify"
	"github.com/iotaledger/goshimmer/packages/unbreakable_consensus/social_consensus"
)

type Epoch struct {
	onlineMana map[string]*social_consensus.Node
	number     int
}

func NewEpoch(number int) *Epoch {
	return &Epoch{
		onlineMana: make(map[string]*social_consensus.Node),
		number:     number,
	}
}

func (epoch *Epoch) GetNumber() int {
	return epoch.number
}

func (epoch *Epoch) String() string {
	return stringify.Struct("Epoch",
		stringify.StructField("number", epoch.number),
	)
}
