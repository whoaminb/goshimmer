package ledgerstate

type TransferOutputReference struct {
	addressHash  AddressHash
	transferHash TransferHash
}

func NewTransferOutputReference(addressHash AddressHash, transferHash TransferHash) *TransferOutputReference {
	return &TransferOutputReference{
		addressHash:  addressHash,
		transferHash: transferHash,
	}
}
