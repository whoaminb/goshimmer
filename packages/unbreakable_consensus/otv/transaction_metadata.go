package social_consensus

type TransactionMetadata struct {
	solid                   bool
	transactionLocallyLiked bool
	realityLocallyLiked     bool
	reality                 int
}

func NewTransactionMetadata() *TransactionMetadata {
	return &TransactionMetadata{}
}

func (transactionMetadata *TransactionMetadata) SetSolid(solid bool) {
	transactionMetadata.solid = solid
}

func (transactionMetadata *TransactionMetadata) IsSolid() bool {
	return transactionMetadata.solid
}

func (transactionMetadata *TransactionMetadata) SetTransactionLocallyLiked(liked bool) {
	transactionMetadata.transactionLocallyLiked = liked
}

func (transactionMetadata *TransactionMetadata) IsTransactionLocallyLiked() bool {
	return transactionMetadata.transactionLocallyLiked
}

func (transactionMetadata *TransactionMetadata) SetRealityLocallyLiked(liked bool) {
	transactionMetadata.realityLocallyLiked = liked
}

func (transactionMetadata *TransactionMetadata) IsRealityLocallyLiked() bool {
	return transactionMetadata.realityLocallyLiked
}

func (transactionMetadata *TransactionMetadata) SetReality(reality int) {
	transactionMetadata.reality = reality
}

func (transactionMetadata *TransactionMetadata) GetReality() int {
	return transactionMetadata.reality
}
