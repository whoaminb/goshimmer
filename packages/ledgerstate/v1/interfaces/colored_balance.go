package interfaces

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/v1/hash"
)

type ColoredBalance interface {
	GetColor() hash.Color
	GetValue() uint64
	String() string
}
