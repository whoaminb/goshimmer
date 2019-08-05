package mana

import (
	"fmt"
)

type Balance struct {
	calculator                 *Calculator
	currentBalance             uint64
	lastErosion                uint64
	roundingErrorInLastErosion float64
}

func NewBalance(calculator *Calculator) *Balance {
	return &Balance{
		calculator:                 calculator,
		currentBalance:             0,
		lastErosion:                0,
		roundingErrorInLastErosion: 0,
	}
}

func (balance *Balance) GetValue() uint64 {
	return balance.currentBalance
}

func (balance *Balance) AddTransfer(movedCoins uint64, receivedTime uint64, spentTime uint64) {
	gainedMana, lastErosion, _ := balance.calculator.ManaOfTransferDiscrete(movedCoins, receivedTime, spentTime)

	if lastErosion >= balance.lastErosion {
		balance.Erode(lastErosion)
	} else {
		fmt.Println("empty")
		// revert old actions
		// apply new
		// replay old
	}

	balance.currentBalance += gainedMana
}

func (balance *Balance) Erode(erosionTime uint64) {
	if balance.lastErosion <= erosionTime {
		balance.currentBalance, balance.lastErosion, _ = balance.calculator.ErodedManaDiscrete(balance.currentBalance, balance.lastErosion, erosionTime)
	} else {
		fmt.Println("empty")
		// revert old erosions
	}
}
