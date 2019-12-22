package conflict

import "github.com/iotaledger/goshimmer/packages/binary/types"

type IdList []Id

func (idList IdList) Clone() (clone IdList) {
	clone = make(IdList, len(idList))

	for key, value := range idList {
		clone[key] = value
	}

	return
}

func (idList IdList) ToSet() (set IdSet) {
	set = make(IdSet)

	for _, value := range idList {
		set[value] = types.Void
	}

	return
}
