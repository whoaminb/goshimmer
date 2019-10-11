package ledgerstate

type AddressHash = string

type TransferHash = string

type ColorHash = string

type TransferOutputs map[TransferHash]*TransferOutput

func (transferOutputs TransferOutputs) String() (result string) {
	for _, transferOutput := range transferOutputs {
		result += transferOutput.String() + " "
	}

	return
}
