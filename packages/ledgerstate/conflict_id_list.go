package ledgerstate

type ConflictIdList []ConflictId

func (conflictIdList ConflictIdList) Clone() (clone ConflictIdList) {
	clone = make(ConflictIdList, len(conflictIdList))

	for key, value := range conflictIdList {
		clone[key] = value
	}

	return
}

func (conflictIdList ConflictIdList) ToSet() (set ConflictIdSet) {
	set = make(ConflictIdSet)

	for _, value := range conflictIdList {
		set[value] = void
	}

	return
}
