package valuetangle

import (
	"fmt"
	"testing"
	"time"

	"github.com/mr-tron/base58"

	"github.com/iotaledger/goshimmer/packages/binary/address"
	"github.com/iotaledger/goshimmer/packages/binary/transactionmetadata"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/coloredcoins"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"

	"github.com/iotaledger/goshimmer/packages/binary/signature/ed25119"

	"github.com/iotaledger/goshimmer/packages/binary/transaction/payload/valuetransfer"

	"github.com/iotaledger/goshimmer/packages/binary/identity"
	"github.com/iotaledger/goshimmer/packages/binary/transaction"

	"github.com/iotaledger/goshimmer/packages/binary/tangle"
	"github.com/iotaledger/goshimmer/packages/binary/valuetangle/model"
	"github.com/iotaledger/hive.go/events"
)

var localSnapshot *tangle.Snapshot
var keyPairOfAddress1 = ed25119.GenerateKeyPair()
var keyPairOfAddress2 = ed25119.GenerateKeyPair()

func getLocalSnapshot() *tangle.Snapshot {
	if localSnapshot == nil {
		localSnapshot = &tangle.Snapshot{
			SolidEntryPoints: map[transaction.Id]map[address.Address]*coloredcoins.ColoredBalance{
				transaction.NewId([]byte("tx0")): nil,
				transaction.NewId([]byte("tx1")): {
					address.FromPublicKey(keyPairOfAddress1.PublicKey): coloredcoins.NewColoredBalance(coloredcoins.NewColor("IOTA"), 1337),
					address.FromPublicKey(keyPairOfAddress2.PublicKey): coloredcoins.NewColoredBalance(coloredcoins.NewColor("IOTA"), 1337),
				},
				transaction.NewId([]byte("tx2")): {
					address.New([]byte("address3")): coloredcoins.NewColoredBalance(coloredcoins.NewColor("IOTA"), 1337),
					address.New([]byte("address4")): coloredcoins.NewColoredBalance(coloredcoins.NewColor("IOTA"), 1337),
				},
			},
		}
	}

	return localSnapshot
}

func TestValueTangle(t *testing.T) {
	transactionTangle := tangle.New([]byte("testtangle"))
	transactionTangle.Prune()

	valueTangle := New(transactionTangle)
	valueTangle.Prune()

	transactionTangle.Events.TransactionAttached.Attach(events.NewClosure(func(cachedTransaction *transaction.CachedTransaction, cachedTransactionMetadata *transactionmetadata.CachedTransactionMetadata) {
		transactionId := cachedTransaction.Unwrap().GetId()

		if cachedTransaction.Unwrap().GetPayload().GetType() != valuetransfer.Type {
			fmt.Println("[TRANSACTION TANGLE] Data transaction attached: ", base58.Encode(transactionId[:]))
		} else {
			fmt.Println("[TRANSACTION TANGLE] Value Transaction attached:", base58.Encode(transactionId[:]))
		}
	}))

	transactionTangle.Events.TransactionSolid.Attach(events.NewClosure(func(cachedTransaction *transaction.CachedTransaction, cachedTransactionMetadata *transactionmetadata.CachedTransactionMetadata) {
		transactionId := cachedTransaction.Unwrap().GetId()

		if cachedTransaction.Unwrap().GetPayload().GetType() != valuetransfer.Type {
			fmt.Println("[TRANSACTION TANGLE] Data transaction solid:    ", base58.Encode(transactionId[:]))
		} else {
			fmt.Println("[TRANSACTION TANGLE] Value Transaction solid:   ", base58.Encode(transactionId[:]))
		}
	}))

	valueTangle.Events.TransferAttached.Attach(events.NewClosure(func(cachedValueTransfer *model.CachedValueTransfer, cachedTransferMetadata *model.CachedTransferMetadata) {
		transactionId := cachedValueTransfer.Unwrap().GetId()

		fmt.Println("[VALUE TANGLE]       Value Transaction attached:", base58.Encode(transactionId[:]))
	}))
	valueTangle.Events.TransferSolid.Attach(events.NewClosure(func(cachedValueTransfer *model.CachedValueTransfer, cachedTransferMetadata *model.CachedTransferMetadata) {
		transactionId := cachedValueTransfer.Unwrap().GetId()

		fmt.Println("[VALUE TANGLE]       Value Transaction solid:   ", base58.Encode(transactionId[:]))
	}))

	transactionTangle.LoadSnapshot(getLocalSnapshot())

	myIdentity := identity.Generate()

	transactionTangle.AttachTransaction(transaction.New(transaction.EmptyId, transaction.EmptyId, myIdentity, valuetransfer.New().
		AddInput(transfer.NewId([]byte("tx1")), address.FromPublicKey(keyPairOfAddress1.PublicKey)).
		AddOutput(address.FromPublicKey(keyPairOfAddress2.PublicKey), coloredcoins.NewColoredBalance(coloredcoins.NewColor("IOTA"), 12)).
		Sign(keyPairOfAddress1)))

	time.Sleep(1 * time.Second)

	valueTangle.Shutdown()
}
