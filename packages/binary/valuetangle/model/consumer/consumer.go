package consumer

import (
	"github.com/iotaledger/goshimmer/packages/binary/valuetangle/model/transferoutput"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/transfer"
	"github.com/iotaledger/hive.go/objectstorage"
)

type Consumer struct {
	objectstorage.StorableObjectFlags

	transferOutputId   transferoutput.Id
	consumerTransferId transfer.Id
}

func New(transferOutput transferoutput.Id, consumerTransfer transfer.Id) *Consumer {
	return &Consumer{
		transferOutputId:   transferOutput,
		consumerTransferId: consumerTransfer,
	}
}

func (consumer *Consumer) GetTransferOutputId() transferoutput.Id {
	return consumer.transferOutputId
}

func (consumer *Consumer) GetConsumerTransferId() transfer.Id {
	return consumer.consumerTransferId
}

var _ objectstorage.StorableObject = &Consumer{}
