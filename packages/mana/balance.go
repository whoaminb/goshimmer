package mana

import (
	"fmt"
)

type Balance struct {
	calculator               *Calculator
	currentBalance           uint64
	lastErosion              uint64
	accumulatedRoundingError float64
}

func NewBalance(calculator *Calculator) *Balance {
	return &Balance{
		calculator:               calculator,
		currentBalance:           0,
		lastErosion:              0,
		accumulatedRoundingError: 0,
	}
}

func (balance *Balance) GetValue() uint64 {
	return balance.currentBalance
}

func (balance *Balance) AddTransfer(movedCoins uint64, receivedTime uint64, spentTime uint64) {
	gainedMana, roundingError := balance.calculator.GenerateMana(movedCoins, spentTime-receivedTime)

	if balance.currentBalance != 0 {
		if spentTime >= balance.lastErosion {
			balance.Erode(spentTime)
		} else {
			fmt.Println("empty")
			// revert old actions
			// apply new
			// replay old
		}
	}

	balance.currentBalance += gainedMana
	balance.accumulatedRoundingError += roundingError
	balance.lastErosion = spentTime

	fmt.Println("GENERATE: ", spentTime-receivedTime, movedCoins, gainedMana)
}

func (balance *Balance) Erode(erosionTime uint64) {
	if balance.lastErosion <= erosionTime {
		erodedMana, _ := balance.calculator.ErodeMana(balance.currentBalance, erosionTime-balance.lastErosion)

		fmt.Println("ERODE: ", erosionTime-balance.lastErosion, balance.currentBalance, erodedMana)

		balance.currentBalance = erodedMana
	} else {
		fmt.Println("empty")
		// revert old erosions
	}
}
