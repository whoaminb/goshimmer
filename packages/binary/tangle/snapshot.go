package tangle

import (
	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/coloredcoins"
)

type Snapshot struct {
	SolidEntryPoints map[transaction.Id]map[address.Address]*coloredcoins.ColoredBalance
}
