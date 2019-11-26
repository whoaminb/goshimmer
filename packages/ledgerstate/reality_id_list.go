package ledgerstate

type RealityIdList []RealityId

func (realityIdList RealityIdList) Clone() (clone RealityIdList) {
	clone = make(RealityIdList, len(realityIdList))

	for key, value := range realityIdList {
		clone[key] = value
	}

	return
}

func (realityIdList RealityIdList) ToSet() (set RealityIdSet) {
	set = make(RealityIdSet)

	for _, value := range realityIdList {
		set[value] = void
	}

	return
}
