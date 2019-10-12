package ledgerstate

type AddressHash = string

type TransferHash = string

type ColorHash = string

type TransferOutputs map[TransferHash]*TransferOutput

func (transferOutputs TransferOutputs) String() (result string) {
	result = "TransferOutputs [[[[[[[[[["

	if len(transferOutputs) >= 1 {
		result += "\n"

		for _, transferOutput := range transferOutputs {
			result += transferOutput.String() + "\n"
		}
	}

	result += "]]]]]]]]]]"

	return
}
