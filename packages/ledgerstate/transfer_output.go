package ledgerstate

// region TransferOutput ///////////////////////////////////////////////////////////////////////////////////////////////

type TransferOutput struct {
	ledgerState     *LedgerState
	realityHash     RealityId
	addressHash     AddressHash
	transferHash    TransferHash
	coloredBalances map[Color]*ColoredBalance
	consumers       []TransferHash
}

func NewTransferOutput(ledgerState *LedgerState, realityHash RealityId, addressHash AddressHash, transferHash TransferHash, coloredBalances ...*ColoredBalance) (result *TransferOutput) {
	result = &TransferOutput{
		ledgerState:     ledgerState,
		addressHash:     addressHash,
		transferHash:    transferHash,
		realityHash:     realityHash,
		coloredBalances: make(map[Color]*ColoredBalance),
		consumers:       make([]TransferHash, 0),
	}

	for _, balance := range coloredBalances {
		result.coloredBalances[balance.GetColor()] = balance
	}

	return
}

func (transferOutput *TransferOutput) GetRealityId() RealityId {
	return transferOutput.realityHash
}

func (transferOutput *TransferOutput) GetReality(realityId RealityId) *Reality {
	return transferOutput.ledgerState.GetReality(realityId)
}

func (transferOutput *TransferOutput) GetAddressHash() AddressHash {
	return transferOutput.addressHash
}

func (transferOutput *TransferOutput) GetTransferHash() TransferHash {
	return transferOutput.transferHash
}

func (transferOutput *TransferOutput) GetColoredBalances() map[Color]*ColoredBalance {
	return transferOutput.coloredBalances
}

func (transferOutput *TransferOutput) GetConsumers() []TransferHash {
	return transferOutput.consumers
}

func (transferOutput *TransferOutput) Exists() bool {
	return transferOutput != nil
}

func (transferOutput *TransferOutput) String() (result string) {
	result = "TransferOutput {\n"
	result += "    RealityHash:  \"" + string(transferOutput.realityHash) + "\",\n"
	result += "    AddressHash:  \"" + string(transferOutput.addressHash) + "\",\n"
	result += "    TransferHash: \"" + string(transferOutput.transferHash) + "\",\n"

	for _, coloredBalance := range transferOutput.coloredBalances {
		result += "    " + coloredBalance.String() + ",\n"
	}

	result += "}"

	return
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region TransferOutputReference //////////////////////////////////////////////////////////////////////////////////////

type TransferOutputReference struct {
	realityId    RealityId
	addressHash  AddressHash
	transferHash TransferHash
}

func NewTransferOutputReference(realityId RealityId, addressHash AddressHash, transferHash TransferHash) *TransferOutputReference {
	return &TransferOutputReference{
		realityId:    realityId,
		addressHash:  addressHash,
		transferHash: transferHash,
	}
}

func (reference *TransferOutputReference) GetRealityId() RealityId {
	return reference.realityId
}

func (reference *TransferOutputReference) GetAddressHash() AddressHash {
	return reference.addressHash
}

func (reference *TransferOutputReference) GetTransferHash() TransferHash {
	return reference.transferHash
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
