package ledgerstate

type TransferOutputReferenceOld struct {
	realityId    RealityId
	addressHash  AddressHash
	transferHash TransferHash
}

func NewTransferOutputReferenceOld(realityId RealityId, addressHash AddressHash, transferHash TransferHash) *TransferOutputReferenceOld {
	return &TransferOutputReferenceOld{
		realityId:    realityId,
		addressHash:  addressHash,
		transferHash: transferHash,
	}
}

func (reference *TransferOutputReferenceOld) GetRealityId() RealityId {
	return reference.realityId
}

func (reference *TransferOutputReferenceOld) GetAddressHash() AddressHash {
	return reference.addressHash
}

func (reference *TransferOutputReferenceOld) GetTransferHash() TransferHash {
	return reference.transferHash
}
