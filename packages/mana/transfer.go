package mana

type Transfer struct {
	movedCoins   uint64
	receivedTime uint64
	spentTime    uint64
}

func NewTransfer(movedCoins uint64, receivedTime uint64, spentTime uint64) *Transfer {
	return &Transfer{
		movedCoins:   movedCoins,
		receivedTime: receivedTime,
		spentTime:    spentTime,
	}
}
