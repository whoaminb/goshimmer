package ledgerstate

type ConflictIdSet map[ConflictId]empty

func NewConflictIdSet(conflictIds ...ConflictId) (conflictIdSet ConflictIdSet) {
	conflictIdSet = make(ConflictIdSet)

	for _, realityId := range conflictIds {
		conflictIdSet[realityId] = void
	}

	return
}

func (conflictIdSet ConflictIdSet) Clone() (clone ConflictIdSet) {
	clone = make(ConflictIdSet, len(conflictIdSet))

	for key := range conflictIdSet {
		clone[key] = void
	}

	return
}

func (conflictIdSet ConflictIdSet) ToList() (list ConflictIdList) {
	list = make(ConflictIdList, len(conflictIdSet))

	i := 0
	for key := range conflictIdSet {
		list[i] = key

		i++
	}

	return
}
