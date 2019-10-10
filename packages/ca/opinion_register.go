package ca

type OpinionRegister struct {
	pendingOpinions map[string]*Opinion
	appliedOpinions map[string]*Opinion
}

func NewOpinionRegister() *OpinionRegister {
	return &OpinionRegister{
		pendingOpinions: make(map[string]*Opinion),
		appliedOpinions: make(map[string]*Opinion),
	}
}

func (opinionRegister *OpinionRegister) GetPendingOpinions() map[string]*Opinion {
	return opinionRegister.pendingOpinions
}

func (opinionRegister *OpinionRegister) GetAppliedOpinions() map[string]*Opinion {
	return opinionRegister.appliedOpinions
}

func (opinionRegister *OpinionRegister) GetOpinion(transactionId string) (opinion *Opinion) {
	if changedOpinion := opinionRegister.pendingOpinions[transactionId]; changedOpinion.Exists() {
		opinion = changedOpinion
	} else {
		opinion = opinionRegister.appliedOpinions[transactionId]
	}

	return
}

func (opinionRegister *OpinionRegister) CreateOpinion(transactionId string) (opinion *Opinion) {
	if opinion = opinionRegister.GetOpinion(transactionId); opinion.Exists() && opinion.IsPending() {
		return
	}

	opinion = NewOpinion()
	opinion.SetInitial(true)
	opinion.SetPending(true)

	opinionRegister.pendingOpinions[transactionId] = opinion

	return opinion
}

func (opinionRegister *OpinionRegister) ApplyPendingOpinions() {
	for transactionId, opinion := range opinionRegister.pendingOpinions {
		opinion.SetPending(false)

		opinionRegister.appliedOpinions[transactionId] = opinion
	}

	opinionRegister.pendingOpinions = make(map[string]*Opinion)
}
