package ledgerstate

type TransferOutput struct {
	hash            TransferHash
	coloredBalances map[ColorHash]*ColoredBalance
}

func NewTransferOutput(hash TransferHash) *TransferOutput {
	return &TransferOutput{
		hash:            hash,
		coloredBalances: make(map[ColorHash]*ColoredBalance),
	}
}

func (transferOutput *TransferOutput) SetColoredBalance(color ColorHash, balance uint64) *TransferOutput {
	transferOutput.coloredBalances[color] = NewColoredBalance(color, balance)

	return transferOutput
}

func (transferOutput *TransferOutput) GetHash() TransferHash {
	return transferOutput.hash
}

func (transferOutput *TransferOutput) Exists() bool {
	return transferOutput != nil
}

func (transferOutput *TransferOutput) String() (result string) {
	result = "TransferOutput(" + transferOutput.hash + ") {"

	if len(transferOutput.coloredBalances) >= 1 {
		for _, coloredBalance := range transferOutput.coloredBalances {
			result += "\n    " + coloredBalance.String()
		}

		result += "\n"
	}

	result += "}"

	return
}
