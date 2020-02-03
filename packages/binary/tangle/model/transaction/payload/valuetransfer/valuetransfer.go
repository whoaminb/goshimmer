package valuetransfer

import (
	"encoding/binary"
	"sync"

	"github.com/iotaledger/goshimmer/packages/binary/signature/ed25119"
	"github.com/iotaledger/goshimmer/packages/binary/tangle/model/transaction/payload"
	"github.com/iotaledger/goshimmer/packages/binary/types"
	"github.com/iotaledger/goshimmer/packages/binary/valuetangle/model/address"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/coloredcoins"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
)

type ValueTransfer struct {
	inputs         []*transfer.OutputReference
	outputs        map[address.Address][]*coloredcoins.ColoredBalance
	signatures     map[ed25119.PublicKey]ed25119.Signature
	payloadBytes   []byte
	signatureBytes []byte

	inputsMutex         sync.RWMutex
	outputsMutex        sync.RWMutex
	signaturesMutex     sync.RWMutex
	payloadBytesMutex   sync.RWMutex
	signatureBytesMutex sync.RWMutex
}

var Type = payload.Type(1)

func New() *ValueTransfer {
	return &ValueTransfer{
		inputs:     make([]*transfer.OutputReference, 0),
		outputs:    make(map[address.Address][]*coloredcoins.ColoredBalance),
		signatures: make(map[ed25119.PublicKey]ed25119.Signature),
	}
}

func (valueTransfer *ValueTransfer) GetType() payload.Type {
	return Type
}

func (valueTransfer *ValueTransfer) AddInput(transferId transfer.Id, address address.Address) *ValueTransfer {
	if valueTransfer.isFinalized() {
		panic("you can not add inputs after you have started finalizing (sign / marshal) the ValueTransfer")
	}

	valueTransfer.inputsMutex.Lock()
	valueTransfer.inputs = append(valueTransfer.inputs, transfer.NewOutputReference(transferId, address))
	valueTransfer.inputsMutex.Unlock()

	return valueTransfer
}

func (valueTransfer *ValueTransfer) GetInputs() (result []*transfer.OutputReference) {
	valueTransfer.inputsMutex.RLock()
	result = valueTransfer.inputs
	valueTransfer.inputsMutex.RUnlock()

	return
}

func (valueTransfer *ValueTransfer) AddOutput(address address.Address, balance *coloredcoins.ColoredBalance) *ValueTransfer {
	if valueTransfer.isFinalized() {
		panic("you can not add outputs after you have started finalizing (sign / marshal) the ValueTransfer")
	}

	valueTransfer.outputsMutex.Lock()
	if _, addressExists := valueTransfer.outputs[address]; !addressExists {
		valueTransfer.outputs[address] = make([]*coloredcoins.ColoredBalance, 0)
	}

	valueTransfer.outputs[address] = append(valueTransfer.outputs[address], balance)
	valueTransfer.outputsMutex.Unlock()

	return valueTransfer
}

func (valueTransfer *ValueTransfer) GetOutputs() (result map[address.Address][]*coloredcoins.ColoredBalance) {
	valueTransfer.outputsMutex.RLock()
	result = valueTransfer.outputs
	valueTransfer.outputsMutex.RUnlock()

	return
}

func (valueTransfer *ValueTransfer) Sign(keyPair ed25119.KeyPair) *ValueTransfer {
	payloadBytes := valueTransfer.marshalPayloadBytes()

	valueTransfer.signaturesMutex.RLock()
	if _, signatureExists := valueTransfer.signatures[keyPair.PublicKey]; !signatureExists {
		valueTransfer.signaturesMutex.RUnlock()

		valueTransfer.signaturesMutex.Lock()
		if _, signatureExists := valueTransfer.signatures[keyPair.PublicKey]; !signatureExists {
			valueTransfer.signatures[keyPair.PublicKey] = keyPair.PrivateKey.Sign(payloadBytes)
		}
		valueTransfer.signaturesMutex.Unlock()
	} else {
		valueTransfer.signaturesMutex.RUnlock()
	}

	return valueTransfer
}

func (valueTransfer *ValueTransfer) VerifySignatures() bool {
	_, _ = valueTransfer.MarshalBinary()
	payloadBytes := valueTransfer.marshalPayloadBytes()

	addressesToSign := make(map[address.Address]types.Empty)
	for _, input := range valueTransfer.inputs {
		addressesToSign[input.GetAddress()] = types.Void
	}

	for publicKey, signature := range valueTransfer.signatures {
		if publicKey.VerifySignature(payloadBytes, signature) {
			delete(addressesToSign, address.FromPublicKey(publicKey))
		} else {
			panic("INVALID SIGNATURE")
		}
	}

	return len(addressesToSign) == 0
}

func (valueTransfer *ValueTransfer) MarshalBinary() (result []byte, err error) {
	payloadBytes := valueTransfer.marshalPayloadBytes()
	signatureBytes := valueTransfer.marshalSignatureBytes(payloadBytes)

	result = append(payloadBytes, signatureBytes...)

	return
}

func (valueTransfer *ValueTransfer) UnmarshalBinary(bytes []byte) (err error) {
	offset := 0

	payloadBytesOffset := offset
	if err = valueTransfer.unmarshalInputs(bytes, &offset); err != nil {
		return
	}

	if err = valueTransfer.unmarshalOutputs(bytes, &offset); err != nil {
		return
	}

	payloadBytesLength := offset - payloadBytesOffset
	valueTransfer.payloadBytes = make([]byte, payloadBytesLength)
	copy(valueTransfer.payloadBytes, bytes[payloadBytesOffset:payloadBytesOffset+payloadBytesLength])

	signatureBytesOffset := offset
	if err = valueTransfer.unmarshalSignatures(bytes, &offset); err != nil {
		return
	}

	signatureBytesLength := offset - signatureBytesOffset
	valueTransfer.signatureBytes = make([]byte, signatureBytesLength)
	copy(valueTransfer.signatureBytes, bytes[signatureBytesOffset:signatureBytesOffset+signatureBytesLength])

	return
}

func (valueTransfer *ValueTransfer) isFinalized() (result bool) {
	valueTransfer.payloadBytesMutex.RLock()
	result = valueTransfer.payloadBytes != nil
	valueTransfer.payloadBytesMutex.RUnlock()

	return
}

func (valueTransfer *ValueTransfer) marshalPayloadBytes() (result []byte) {
	valueTransfer.payloadBytesMutex.RLock()
	if valueTransfer.payloadBytes == nil {
		valueTransfer.payloadBytesMutex.RUnlock()

		valueTransfer.payloadBytesMutex.Lock()
		if valueTransfer.payloadBytes == nil {
			result = append(valueTransfer.marshalInputs(), valueTransfer.marshalOutputs()...)

			valueTransfer.payloadBytes = result
		} else {
			result = valueTransfer.payloadBytes
		}
		valueTransfer.payloadBytesMutex.Unlock()
	} else {
		result = valueTransfer.payloadBytes

		valueTransfer.payloadBytesMutex.RUnlock()
	}

	return
}

func (valueTransfer *ValueTransfer) marshalInputs() (result []byte) {
	inputCount := len(valueTransfer.inputs)
	marshaledTransferOutputReferenceLength := transfer.IdLength + address.Length

	result = make([]byte, 4+inputCount*marshaledTransferOutputReferenceLength)
	offset := 0

	binary.LittleEndian.PutUint32(result[offset:], uint32(inputCount))
	offset += 4

	for _, transferOutputReference := range valueTransfer.inputs {
		if marshaledTransferOutputReference, err := transferOutputReference.MarshalBinary(); err != nil {
			panic(err)
		} else {
			copy(result[offset:], marshaledTransferOutputReference)
			offset += marshaledTransferOutputReferenceLength
		}
	}

	return
}

func (valueTransfer *ValueTransfer) marshalOutputs() (result []byte) {
	totalLength := 4
	for _, outputs := range valueTransfer.outputs {
		totalLength += address.Length

		totalLength += 4

		for range outputs {
			totalLength += coloredcoins.ColorLength + 8
		}
	}

	result = make([]byte, totalLength)
	offset := 0

	binary.LittleEndian.PutUint32(result[offset:], uint32(len(valueTransfer.outputs)))
	offset += 4

	for outputAddress, outputs := range valueTransfer.outputs {
		copy(result[offset:], outputAddress[:])
		offset += address.Length

		binary.LittleEndian.PutUint32(result[offset:], uint32(len(outputs)))
		offset += 4

		for _, coloredBalance := range outputs {
			if marshaledColoredBalance, marshalErr := coloredBalance.MarshalBinary(); marshalErr != nil {
				panic(marshalErr)
			} else {
				copy(result[offset:], marshaledColoredBalance)
				offset += coloredcoins.ColorLength + 8
			}
		}
	}

	return
}

func (valueTransfer *ValueTransfer) marshalSignatureBytes(payloadBytes []byte) (result []byte) {
	valueTransfer.signatureBytesMutex.RLock()
	if valueTransfer.signatureBytes == nil {
		valueTransfer.signatureBytesMutex.RUnlock()

		valueTransfer.signatureBytesMutex.Lock()
		if valueTransfer.signatureBytes == nil {
			signatureCount := len(valueTransfer.signatures)
			result = make([]byte, 4+signatureCount*(ed25119.PublicKeySize+ed25119.SignatureSize))
			offset := 0

			binary.LittleEndian.PutUint32(result[offset:], uint32(signatureCount))
			offset += 4

			valueTransfer.signaturesMutex.RLock()
			for publicKey, signature := range valueTransfer.signatures {
				copy(result[offset:], publicKey[:])
				offset += ed25119.PublicKeySize

				copy(result[offset:], signature[:])
				offset += ed25119.SignatureSize
			}
			valueTransfer.signaturesMutex.RUnlock()

			valueTransfer.signatureBytes = result
		} else {
			result = valueTransfer.signatureBytes
		}
		valueTransfer.signatureBytesMutex.Unlock()
	} else {
		result = valueTransfer.signatureBytes

		valueTransfer.signatureBytesMutex.RUnlock()
	}

	return
}

func (valueTransfer *ValueTransfer) unmarshalInputs(bytes []byte, offset *int) (err error) {
	inputCount := int(binary.LittleEndian.Uint32(bytes[*offset:]))
	*offset += 4

	valueTransfer.inputs = make([]*transfer.OutputReference, inputCount)
	marshaledTransferOutputReferenceLength := transfer.IdLength + address.Length
	for i := 0; i < inputCount; i++ {
		var transferOutputReference transfer.OutputReference
		if err = transferOutputReference.UnmarshalBinary(bytes[*offset:]); err != nil {
			return
		}
		*offset += marshaledTransferOutputReferenceLength

		valueTransfer.inputs[i] = &transferOutputReference
	}

	return
}

func (valueTransfer *ValueTransfer) unmarshalOutputs(bytes []byte, offset *int) (err error) {
	outputCount := int(binary.LittleEndian.Uint32(bytes[*offset:]))
	*offset += 4

	valueTransfer.outputs = make(map[address.Address][]*coloredcoins.ColoredBalance)

	for i := 0; i < outputCount; i++ {
		var outputAddress address.Address
		if err = outputAddress.UnmarshalBinary(bytes[*offset:]); err != nil {
			return
		}
		*offset += address.Length

		outputsCount := int(binary.LittleEndian.Uint32(bytes[*offset:]))
		*offset += 4

		valueTransfer.outputs[outputAddress] = make([]*coloredcoins.ColoredBalance, outputsCount)

		for j := 0; j < outputsCount; j++ {
			var coloredBalance coloredcoins.ColoredBalance
			if err = coloredBalance.UnmarshalBinary(bytes[*offset:]); err != nil {
				return
			}
			*offset += coloredcoins.BalanceLength

			valueTransfer.outputs[outputAddress][j] = &coloredBalance
		}
	}

	return
}

func (valueTransfer *ValueTransfer) unmarshalSignatures(bytes []byte, offset *int) (err error) {
	signatureCount := int(binary.LittleEndian.Uint32(bytes[*offset:]))
	*offset += 4

	valueTransfer.signatures = make(map[ed25119.PublicKey]ed25119.Signature)

	for i := 0; i < signatureCount; i++ {
		var publicKey ed25119.PublicKey
		if err = publicKey.UnmarshalBinary(bytes[*offset:]); err != nil {
			return
		}
		*offset += ed25119.PublicKeySize

		var signature ed25119.Signature
		if err = signature.UnmarshalBinary(bytes[*offset:]); err != nil {
			return
		}
		*offset += ed25119.SignatureSize

		valueTransfer.signatures[publicKey] = signature
	}

	return
}

func init() {
	payload.RegisterType(Type, func(data []byte) (payload payload.Payload, err error) {
		payload = &ValueTransfer{}
		err = payload.UnmarshalBinary(data)

		return
	})
}
