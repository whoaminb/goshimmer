package otv

import (
	"fmt"
)

type ElderMask uint64

func (elderMask ElderMask) Union(otherElderMask ElderMask) ElderMask {
	return elderMask | otherElderMask
}

func (elderMask ElderMask) Contains(otherMask ElderMask) bool {
	return elderMask&otherMask != 0
}

func (elderMask ElderMask) String() string {
	return "ElderMask(" + fmt.Sprintf("%064b", elderMask) + ")"
}
