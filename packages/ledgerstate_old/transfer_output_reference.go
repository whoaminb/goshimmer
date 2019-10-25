package ledgerstate

type TransferOutputReference struct {
	id           [transferHashLength + addressHashLength]byte
	transferHash TransferHash
	addressHash  AddressHash
}

func NewTransferOutputReference(transferHash TransferHash, addressHash AddressHash) (result *TransferOutputReference) {
	result = &TransferOutputReference{
		transferHash: transferHash,
		addressHash:  addressHash,
	}

	copy(result.id[0:], transferHash[:transferHashLength])
	copy(result.id[transferHashLength:], addressHash[:addressHashLength])

	return
}

func (reference *TransferOutputReference) GetId() []byte {
	return reference.id[:]
}

func (reference *TransferOutputReference) GetTransferHash() TransferHash {
	return reference.transferHash
}

func (reference *TransferOutputReference) GetAddressHash() AddressHash {
	return reference.addressHash
}
