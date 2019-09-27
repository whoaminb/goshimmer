package otv

type ConflictSet struct {
	realities map[int]*Reality
}

func NewConflictSet() *ConflictSet {
	return &ConflictSet{
		realities: make(map[int]*Reality),
	}
}

func (conflictSet *ConflictSet) GetRealities() map[int]*Reality {
	return conflictSet.realities
}

func (conflictSet *ConflictSet) GetReality(id int) *Reality {
	return conflictSet.realities[id]
}

func (conflictSet *ConflictSet) AddReality(id int) (result *Reality) {
	if _, exists := conflictSet.realities[id]; !exists {
		result = NewReality(conflictSet, id)

		conflictSet.realities[id] = result
	}

	return
}
