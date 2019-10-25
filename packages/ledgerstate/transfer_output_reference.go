package ledgerstate

type TransferOutputReference struct {
	storageKey   []byte
	transferHash TransferHash
	addressHash  AddressHash
}

func NewTransferOutputReference(transferHash TransferHash, addressHash AddressHash) *TransferOutputReference {
	return &TransferOutputReference{
		storageKey:   append(transferHash[:], addressHash[:]...),
		transferHash: transferHash,
		addressHash:  addressHash,
	}
}

func (transferOutputReference *TransferOutputReference) GetStorageKey() []byte {
	return transferOutputReference.storageKey
}
