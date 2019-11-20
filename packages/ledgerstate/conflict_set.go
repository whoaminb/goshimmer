package ledgerstate

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/iotaledger/goshimmer/packages/stringify"

	"github.com/iotaledger/hive.go/objectstorage"
)

type ConflictSet struct {
	id      ConflictSetId
	members map[RealityId]empty

	storageKey  []byte
	ledgerState *LedgerState

	membersMutex sync.RWMutex
}

func newConflictSet(id ConflictSetId) *ConflictSet {
	result := &ConflictSet{
		id:      id,
		members: make(map[RealityId]empty),

		storageKey: make([]byte, conflictSetIdLength),
	}
	copy(result.storageKey, id[:])

	return result
}

func (conflictSet *ConflictSet) AddMember(realityId RealityId) {
	conflictSet.membersMutex.Lock()

	conflictSet.members[realityId] = void

	conflictSet.membersMutex.Unlock()
}

func (conflictSet *ConflictSet) String() string {
	conflictSet.membersMutex.RLock()
	defer conflictSet.membersMutex.RUnlock()

	return stringify.Struct("ConflictSet",
		stringify.StructField("id", conflictSet.id.String()),
		stringify.StructField("members", conflictSet.members),
	)
}

// region support object storage ///////////////////////////////////////////////////////////////////////////////////////

func (conflictSet *ConflictSet) GetStorageKey() []byte {
	return conflictSet.storageKey
}

func (conflictSet *ConflictSet) Update(other objectstorage.StorableObject) {
	fmt.Println("UPDATE")
}

func (conflictSet *ConflictSet) MarshalBinary() ([]byte, error) {
	conflictSet.membersMutex.RLock()

	offset := 0
	membersCount := len(conflictSet.members)
	result := make([]byte, 4+membersCount*realityIdLength)

	binary.LittleEndian.PutUint32(result[offset:], uint32(membersCount))
	offset += 4

	for realityId := range conflictSet.members {
		copy(result[offset:], realityId[:realityIdLength])
		offset += realityIdLength
	}

	conflictSet.membersMutex.RUnlock()

	return result, nil
}

func (conflictSet *ConflictSet) UnmarshalBinary(serializedObject []byte) error {
	if err := conflictSet.id.UnmarshalBinary(conflictSet.storageKey); err != nil {
		return err
	}

	if members, err := conflictSet.unmarshalMembers(serializedObject); err != nil {
		return err
	} else {
		conflictSet.members = members
	}

	return nil
}

func (conflictSet *ConflictSet) unmarshalMembers(serializedConsumers []byte) (map[RealityId]empty, error) {
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
