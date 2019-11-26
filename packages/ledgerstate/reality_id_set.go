package ledgerstate

type RealityIdSet map[RealityId]empty

func NewRealityIdSet(realityIds ...RealityId) (realityIdSet RealityIdSet) {
	realityIdSet = make(RealityIdSet)

	for _, realityId := range realityIds {
		realityIdSet[realityId] = void
	}

	return
}

func (realityIdSet RealityIdSet) Add(realityId RealityId) RealityIdSet {
	realityIdSet[realityId] = void

	return realityIdSet
}

func (realityIdSet RealityIdSet) Remove(realityId RealityId) RealityIdSet {
	delete(realityIdSet, realityId)

	return realityIdSet
}

func (realityIdSet RealityIdSet) Clone() (clone RealityIdSet) {
	clone = make(RealityIdSet, len(realityIdSet))

	for key := range realityIdSet {
		clone[key] = void
	}

	return
}

func (realityIdSet RealityIdSet) ToList() (list RealityIdList) {
	list = make(RealityIdList, len(realityIdSet))

	i := 0
	for key := range realityIdSet {
		list[i] = key

		i++
	}

	return
}
