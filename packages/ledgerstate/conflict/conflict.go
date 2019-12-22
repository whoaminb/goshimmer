package conflict

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/iotaledger/goshimmer/packages/binary/types"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/reality"
	"github.com/iotaledger/goshimmer/packages/stringify"

	"github.com/iotaledger/hive.go/objectstorage"
)

type Conflict struct {
	objectstorage.StorableObjectFlags

	id      Id
	Members map[reality.Id]types.Empty

	storageKey []byte

	membersMutex sync.RWMutex
}

func New(id Id) *Conflict {
	result := &Conflict{
		id:      id,
		Members: make(map[reality.Id]types.Empty),

		storageKey: make([]byte, IdLength),
	}
	copy(result.storageKey, id[:])

	return result
}

func Factory(key []byte) objectstorage.StorableObject {
	result := &Conflict{
		storageKey: make([]byte, len(key)),
	}
	copy(result.storageKey, key)

	return result
}

func (conflict *Conflict) GetId() Id {
	return conflict.id
}

func (conflict *Conflict) AddReality(realityId reality.Id) {
	conflict.membersMutex.Lock()

	conflict.Members[realityId] = types.Void

	conflict.membersMutex.Unlock()
}

func (conflict *Conflict) String() string {
	conflict.membersMutex.RLock()
	defer conflict.membersMutex.RUnlock()

	return stringify.Struct("Conflict",
		stringify.StructField("id", conflict.id.String()),
		stringify.StructField("members", conflict.Members),
	)
}

// region support object storage ///////////////////////////////////////////////////////////////////////////////////////

func (conflict *Conflict) GetStorageKey() []byte {
	return conflict.storageKey
}

func (conflict *Conflict) Update(other objectstorage.StorableObject) {
	fmt.Println("UPDATE")
}

func (conflict *Conflict) MarshalBinary() ([]byte, error) {
	conflict.membersMutex.RLock()

	offset := 0
	membersCount := len(conflict.Members)
	result := make([]byte, 4+membersCount*reality.IdLength)

	binary.LittleEndian.PutUint32(result[offset:], uint32(membersCount))
	offset += 4

	for realityId := range conflict.Members {
		copy(result[offset:], realityId[:reality.IdLength])
		offset += reality.IdLength
	}

	conflict.membersMutex.RUnlock()

	return result, nil
}

func (conflict *Conflict) UnmarshalBinary(serializedObject []byte) error {
	if err := conflict.id.UnmarshalBinary(conflict.storageKey); err != nil {
		return err
	}

	if members, err := conflict.unmarshalMembers(serializedObject); err != nil {
		return err
	} else {
		conflict.Members = members
	}

	return nil
}

func (conflict *Conflict) unmarshalMembers(serializedConsumers []byte) (map[reality.Id]types.Empty, error) {
	offset := 0

	membersCount := int(binary.LittleEndian.Uint32(serializedConsumers[offset:]))
	offset += 4

	members := make(map[reality.Id]types.Empty, membersCount)
	for i := 0; i < membersCount; i++ {
		realityId := reality.Id{}
		if err := realityId.UnmarshalBinary(serializedConsumers[offset:]); err != nil {
			return nil, err
		}
		offset += reality.IdLength

		members[realityId] = types.Void
	}

	return members, nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
