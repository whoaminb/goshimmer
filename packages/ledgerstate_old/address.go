package ledgerstate

type AddressHash [addressHashLength]byte

type Address struct {
	ledgerState *LedgerState
	realityId   RealityId
	hash        AddressHash
}

func NewAddress(ledgerState *LedgerState, realityId RealityId, hash AddressHash) *Address {
	return &Address{
		ledgerState: ledgerState,
		realityId:   realityId,
		hash:        hash,
	}
}

func (address *Address) GetHash() AddressHash {
	return address.hash
}

func (address *Address) GetRealityId() RealityId {
	return address.realityId
}

func (address *Address) GetReality() *Reality {
	return address.ledgerState.GetReality(address.realityId)
}

func (address *Address) GetUnspentTransferOutputs() (unspentTransferOutputs []*TransferOutput) {
	unspentTransferOutputs = make([]*TransferOutput, 0)

	address.collectUnspentTransferOutputs(address.realityId, &unspentTransferOutputs)
	for ancestorRealityId := range address.GetReality().GetAncestorRealities() {
		address.collectUnspentTransferOutputs(ancestorRealityId, &unspentTransferOutputs)
	}

	return unspentTransferOutputs
}

func (address *Address) collectUnspentTransferOutputs(realityId RealityId, unspentTransferOutputs *[]*TransferOutput) {
	address.ledgerState.ForEachTransferOutput(func(transferOutput *TransferOutput) {
		*unspentTransferOutputs = append(*unspentTransferOutputs, transferOutput)
	}, FilterRealities(realityId), FilterAddresses(address.hash), FilterUnspent())
}

func (address *Address) GetTransferOutputs() map[TransferHash]*TransferOutput {
	return nil
}

func (address *Address) GetBalances() map[Color]uint64 {
	balances := make(map[Color]uint64)

	for _, unspentTransferOutput := range address.GetUnspentTransferOutputs() {
		for colorHash, balance := range unspentTransferOutput.GetColoredBalances() {
			balances[colorHash] += balance.GetValue()
		}
	}

	return balances
}
