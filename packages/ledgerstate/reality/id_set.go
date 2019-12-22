package reality

import "github.com/iotaledger/goshimmer/packages/binary/types"

type IdSet map[Id]types.Empty

func NewIdSet(realityIds ...Id) (realityIdSet IdSet) {
	realityIdSet = make(IdSet)

	for _, realityId := range realityIds {
		realityIdSet[realityId] = types.Void
	}

	return
}

func (realityIdSet IdSet) Contains(realityId Id) bool {
	_, exists := realityIdSet[realityId]

	return exists
}

func (realityIdSet IdSet) Add(realityId Id) IdSet {
	realityIdSet[realityId] = types.Void

	return realityIdSet
}

func (realityIdSet IdSet) Remove(realityId Id) IdSet {
	delete(realityIdSet, realityId)

	return realityIdSet
}

func (realityIdSet IdSet) Clone() (clone IdSet) {
	clone = make(IdSet, len(realityIdSet))

	for key := range realityIdSet {
		clone[key] = types.Void
	}

	return
}

func (realityIdSet IdSet) ToList() (list IdList) {
	list = make(IdList, len(realityIdSet))

	i := 0
	for key := range realityIdSet {
		list[i] = key

		i++
	}

	return
}
