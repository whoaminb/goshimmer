package transaction

import (
	"sync"
)

type PayloadUnmarshaler func(data []byte) (Payload, error)

var (
	payloadTypeRegister      map[PayloadType]PayloadUnmarshaler
	payloadTypeRegisterMutex sync.RWMutex
)

func RegisterPayloadType(payloadType PayloadType, unmarshaler PayloadUnmarshaler) {
	payloadTypeRegisterMutex.Lock()
	payloadTypeRegister[payloadType] = unmarshaler
	payloadTypeRegisterMutex.Unlock()
}

func GetPayloadUnmarshaler(payloadType PayloadType) PayloadUnmarshaler {
	payloadTypeRegisterMutex.RLock()
	if unmarshaler, exists := payloadTypeRegister[payloadType]; exists {
		payloadTypeRegisterMutex.RUnlock()

		return unmarshaler
	} else {
		payloadTypeRegisterMutex.RUnlock()

		return createGenericDataPayloadUnmarshaler(payloadType)
	}
}
