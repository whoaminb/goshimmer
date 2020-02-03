package tangle

import (
	"github.com/iotaledger/goshimmer/packages/binary/tangle/model/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/valuetangle/model/address"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/coloredcoins"
)

type Snapshot struct {
	SolidEntryPoints map[transaction.Id]map[address.Address]*coloredcoins.ColoredBalance
}
