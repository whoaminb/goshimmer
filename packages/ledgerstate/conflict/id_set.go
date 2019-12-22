package conflict

import "github.com/iotaledger/goshimmer/packages/binary/types"

type IdSet map[Id]types.Empty

func NewIdSet(conflictIds ...Id) (conflictIdSet IdSet) {
	conflictIdSet = make(IdSet)

	for _, realityId := range conflictIds {
		conflictIdSet[realityId] = types.Void
	}

	return
}

func (idSet IdSet) Clone() (clone IdSet) {
	clone = make(IdSet, len(idSet))

	for key := range idSet {
		clone[key] = types.Void
	}

	return
}

func (idSet IdSet) ToList() (list IdList) {
	list = make(IdList, len(idSet))

	i := 0
	for key := range idSet {
		list[i] = key

		i++
	}

	return
}
