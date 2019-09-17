package unbreakable_consensus

type EpochRegister struct {
	epochs map[int]*Epoch
}

func NewEpochRegister() *EpochRegister {
	return &EpochRegister{
		epochs: make(map[int]*Epoch),
	}
}

func (epochRegister *EpochRegister) GetEpoch(number int) *Epoch {
	if epoch, exists := epochRegister.epochs[number]; exists {
		return epoch
	} else {
		epoch := NewEpoch(number)

		epochRegister.epochs[number] = epoch

		return epoch
	}
}
