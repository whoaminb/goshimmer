package tangle

import (
	"github.com/iotaledger/goshimmer/packages/binary/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/transactionmetadata"
	"github.com/iotaledger/hive.go/events"
)

type Events struct {
	TransactionAttached        *events.Event
	TransactionSolid           *events.Event
	TransactionMissing         *events.Event
	MissingTransactionAttached *events.Event
	TransactionRemoved         *events.Event
}

func newEvents() *Events {
	return &Events{
		TransactionAttached:        events.NewEvent(cachedTransactionEvent),
		TransactionSolid:           events.NewEvent(cachedTransactionEvent),
		TransactionMissing:         events.NewEvent(transactionIdEvent),
		MissingTransactionAttached: events.NewEvent(transactionIdEvent),
		TransactionRemoved:         events.NewEvent(transactionIdEvent),
	}
}

func transactionIdEvent(handler interface{}, params ...interface{}) {
	missingTransactionId := params[0].(transaction.Id)

	handler.(func(transaction.Id))(missingTransactionId)
}

func cachedTransactionEvent(handler interface{}, params ...interface{}) {
	cachedTransaction := params[0].(*transaction.CachedTransaction)
	cachedTransactionMetadata := params[1].(*transactionmetadata.CachedTransactionMetadata)

	cachedTransaction.RegisterConsumer()
	cachedTransactionMetadata.RegisterConsumer()

	handler.(func(*transaction.CachedTransaction, *transactionmetadata.CachedTransactionMetadata))(cachedTransaction, cachedTransactionMetadata)
}
