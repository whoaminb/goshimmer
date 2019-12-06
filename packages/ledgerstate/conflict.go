package ledgerstate

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/iotaledger/goshimmer/packages/stringify"

	"github.com/iotaledger/hive.go/objectstorage"
)

type Conflict struct {
	objectstorage.StorableObjectFlags

	id      ConflictId
	members map[RealityId]empty

	storageKey  []byte
	ledgerState *LedgerState

	membersMutex sync.RWMutex
}

func newConflictSet(id ConflictId) *Conflict {
	result := &Conflict{
		id:      id,
		members: make(map[RealityId]empty),

		storageKey: make([]byte, conflictSetIdLength),
	}
	copy(result.storageKey, id[:])

	return result
}

func (conflict *Conflict) GetId() ConflictId {
	return conflict.id
}

func (conflict *Conflict) AddReality(realityId RealityId) {
	conflict.membersMutex.Lock()

	conflict.members[realityId] = void

	conflict.membersMutex.Unlock()
}

func (conflict *Conflict) String() string {
	conflict.membersMutex.RLock()
	defer conflict.membersMutex.RUnlock()

	return stringify.Struct("Conflict",
		stringify.StructField("id", conflict.id.String()),
		stringify.StructField("members", conflict.members),
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
	membersCount := len(conflict.members)
	result := make([]byte, 4+membersCount*realityIdLength)

	binary.LittleEndian.PutUint32(result[offset:], uint32(membersCount))
	offset += 4

	for realityId := range conflict.members {
		copy(result[offset:], realityId[:realityIdLength])
		offset += realityIdLength
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
		conflict.members = members
	}

	return nil
}

func (conflict *Conflict) unmarshalMembers(serializedConsumers []byte) (map[RealityId]empty, error) {
	offset := 0

	membersCount := int(binary.LittleEndian.Uint32(serializedConsumers[offset:]))
	offset += 4

	members := make(map[RealityId]empty, membersCount)
	for i := 0; i < membersCount; i++ {
		realityId := RealityId{}
		if err := realityId.UnmarshalBinary(serializedConsumers[offset:]); err != nil {
			return nil, err
		}
		offset += realityIdLength

		members[realityId] = void
	}

	return members, nil
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////
