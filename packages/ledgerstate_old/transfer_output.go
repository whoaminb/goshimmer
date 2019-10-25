package ledgerstate

import (
	"github.com/iotaledger/goshimmer/packages/objectstorage"
)

// region TransferOutput ///////////////////////////////////////////////////////////////////////////////////////////////

type TransferOutput struct {
	ledgerState     *LedgerState
	realityId       RealityId
	addressHash     AddressHash
	transferHash    TransferHash
	coloredBalances map[Color]*ColoredBalance
	consumers       []TransferHash
}

func NewTransferOutput(ledgerState *LedgerState, realityId RealityId, addressHash AddressHash, transferHash TransferHash, coloredBalances ...*ColoredBalance) (result *TransferOutput) {
	result = &TransferOutput{
		ledgerState:     ledgerState,
		addressHash:     addressHash,
		transferHash:    transferHash,
		coloredBalances: make(map[Color]*ColoredBalance),
		realityId:       realityId,
		consumers:       make([]TransferHash, 0),
	}

	for _, balance := range coloredBalances {
		result.coloredBalances[balance.GetColor()] = balance
	}

	return
}

func (transferOutput *TransferOutput) GetId() []byte {
	return nil
}

func (transferOutput *TransferOutput) Update(otherObject objectstorage.StorableObject) {
	if otherTransferOutput, ok := otherObject.(*TransferOutput); !ok {
		panic("Update expects the passed in object to be a valid *TransferOutput")
	} else {
		transferOutput.realityId = otherTransferOutput.realityId
	}
}

func (transferOutput *TransferOutput) Marshal() ([]byte, error) {
	return nil, nil
}
func (transferOutput *TransferOutput) Unmarshal(key []byte, serializedObject []byte) (objectstorage.StorableObject, error) {
	return &TransferOutput{}, nil
}

func (transferOutput *TransferOutput) GetRealityId() RealityId {
	return transferOutput.realityId
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
	result += "    RealityHash:  \"" + string(transferOutput.realityId[:]) + "\",\n"
	result += "    AddressHash:  \"" + string(transferOutput.addressHash[:]) + "\",\n"
	result += "    TransferHash: \"" + string(transferOutput.transferHash[:]) + "\",\n"

	for _, coloredBalance := range transferOutput.coloredBalances {
		result += "    " + coloredBalance.String() + ",\n"
	}

	result += "}"

	return
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
