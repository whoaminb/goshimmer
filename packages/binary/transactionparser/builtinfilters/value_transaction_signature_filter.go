package builtinfilters

import (
	"sync"

	"github.com/iotaledger/hive.go/async"

	"github.com/iotaledger/goshimmer/packages/binary/tangle/model/transaction"
	"github.com/iotaledger/goshimmer/packages/binary/tangle/model/transaction/payload/valuetransfer"
)

type ValueTransferSignatureFilter struct {
	onAcceptCallback func(tx *transaction.Transaction)
	onRejectCallback func(tx *transaction.Transaction)
	workerPool       async.WorkerPool

	onAcceptCallbackMutex sync.RWMutex
	onRejectCallbackMutex sync.RWMutex
}

func NewValueTransferSignatureFilter() (result *ValueTransferSignatureFilter) {
	result = &ValueTransferSignatureFilter{}

	return
}

func (filter *ValueTransferSignatureFilter) Filter(tx *transaction.Transaction) {
	filter.workerPool.Submit(func() {
		if payload := tx.GetPayload(); payload.GetType() == valuetransfer.Type {
			if valueTransfer, ok := payload.(*valuetransfer.ValueTransfer); ok && valueTransfer.VerifySignatures() {
				filter.getAcceptCallback()(tx)
			} else {
				filter.getRejectCallback()(tx)
			}
		} else {
			filter.getAcceptCallback()(tx)
		}
	})
}

func (filter *ValueTransferSignatureFilter) OnAccept(callback func(tx *transaction.Transaction)) {
	filter.onAcceptCallbackMutex.Lock()
	filter.onAcceptCallback = callback
	filter.onAcceptCallbackMutex.Unlock()
}

func (filter *ValueTransferSignatureFilter) OnReject(callback func(tx *transaction.Transaction)) {
	filter.onRejectCallbackMutex.Lock()
	filter.onRejectCallback = callback
	filter.onRejectCallbackMutex.Unlock()
}

func (filter *ValueTransferSignatureFilter) Shutdown() {
	filter.workerPool.ShutdownGracefully()
}

func (filter *ValueTransferSignatureFilter) getAcceptCallback() (result func(tx *transaction.Transaction)) {
	filter.onAcceptCallbackMutex.RLock()
	result = filter.onAcceptCallback
	filter.onAcceptCallbackMutex.RUnlock()

	return
}

func (filter *ValueTransferSignatureFilter) getRejectCallback() (result func(tx *transaction.Transaction)) {
	filter.onRejectCallbackMutex.RLock()
	result = filter.onRejectCallback
	filter.onRejectCallbackMutex.RUnlock()

	return
}
