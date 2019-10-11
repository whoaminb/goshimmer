package ledgerstate

type Address struct {
	hash                   AddressHash
	unspentTransferOutputs map[TransferHash]*ColoredBalance
}

func NewAddress(hash AddressHash) *Address {
	return &Address{
		hash: hash,
	}
}

func (address *Address) GetUnspentTransferOutputs() map[TransferHash]*ColoredBalance {
	return address.unspentTransferOutputs
}

func (address *Address) Exists() bool {
	return address != nil
}
