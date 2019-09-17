package social_consensus

import (
	"math/rand"
)

type Society struct {
	elders               []*Node
	totalElderReputation uint64
}

func NewSociety(elders []*Node) (result *Society) {
	result = &Society{}
	result.SetElders(elders)

	return
}

func (society *Society) GetElders() []*Node {
	return society.elders
}

func (society *Society) SetElders(newElders []*Node) {
	society.elders = newElders

	society.totalElderReputation = 0
	for i, elder := range newElders {
		elder.SetElderMask(1 << uint(i))

		society.totalElderReputation += elder.GetReputation()
	}
}

func (society *Society) GetReferencedElderReputation(elderMask ElderMask) (result uint64) {
	for i, elder := range society.GetElders() {
		if elderMask&(1<<uint(i)) > 0 {
			result += elder.GetReputation()
		}
	}

	return
}

func (society *Society) GetTotalElderReputation() uint64 {
	return society.totalElderReputation
}

func (society *Society) GetRandomElder() *Node {
	return society.elders[rand.Intn(len(society.elders))]
}
