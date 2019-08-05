package mana

import (
	"math"
)

type Calculator struct {
	decayInterval   uint64
	decayRate       float64
	coinsPerMana    uint64
	decayFactor     float64
	manaScaleFactor float64
}

func NewCalculator(decayInterval uint64, decayRate float64, coinsPerMana uint64) (result *Calculator) {
	result = &Calculator{
		decayInterval: decayInterval,
		decayRate:     decayRate,
		coinsPerMana:  coinsPerMana,
		decayFactor:   1 - decayRate,
	}

	// make mana reach exactly the token supply
	result.manaScaleFactor = result.decayRate / result.decayFactor

	return
}

func (calculator *Calculator) ManaOfTransferDiscrete(movedCoins uint64, receivedTime uint64, spentTime uint64) (result uint64, lastErosion uint64, roundingError float64) {
	if spentTime <= receivedTime {
		return 0, 0, 0
	}

	baseMana := movedCoins / calculator.coinsPerMana
	scaleFactor := 1 - calculator.decayRate
	erosionIntervals := spentTime/calculator.decayInterval - receivedTime/calculator.decayInterval

	var totalManaReceived float64

	switch true {
	// no decay intervals
	case erosionIntervals == 0:
		lastIntervalDuration := spentTime - receivedTime
		if lastIntervalDuration != 0 {
			totalManaReceived += float64(CoinTimeDestroyed(baseMana, lastIntervalDuration))
		}

		lastErosion = 0

	// only 1 decay interval
	case erosionIntervals == 1:
		firstIntervalDuration := calculator.decayInterval - receivedTime%calculator.decayInterval
		gainsInFirstInterval := float64(CoinTimeDestroyed(baseMana, firstIntervalDuration)) * math.Pow(scaleFactor, float64(erosionIntervals))
		totalManaReceived += gainsInFirstInterval

		lastIntervalDuration := spentTime % calculator.decayInterval
		if lastIntervalDuration != 0 {
			totalManaReceived += float64(CoinTimeDestroyed(baseMana, lastIntervalDuration))
		}

		lastErosion = spentTime - lastIntervalDuration

	// multiple decay intervals
	default:
		firstIntervalDuration := calculator.decayInterval - receivedTime%calculator.decayInterval
		gainsInFirstInterval := float64(CoinTimeDestroyed(baseMana, firstIntervalDuration)) * math.Pow(scaleFactor, float64(erosionIntervals))
		totalManaReceived += gainsInFirstInterval

		gainsInConsecutiveIntervals := float64(CoinTimeDestroyed(baseMana, calculator.decayInterval)) * scaleFactor * (1 - math.Pow(scaleFactor, float64(erosionIntervals-1))) / (1 - scaleFactor)
		totalManaReceived += gainsInConsecutiveIntervals

		lastIntervalDuration := spentTime % calculator.decayInterval
		if lastIntervalDuration != 0 {
			totalManaReceived += float64(CoinTimeDestroyed(baseMana, lastIntervalDuration))
		}

		lastErosion = spentTime - lastIntervalDuration
	}

	result = uint64(totalManaReceived)
	roundingError = totalManaReceived - float64(result)

	return
}

func (calculator *Calculator) ManaOfTransferContinuous(movedCoins uint64, heldTime uint64) (result uint64, roundingError float64) {
	relativeDecayTime := float64(heldTime) / float64(calculator.decayInterval)

	erosionFactor := (1-math.Pow(calculator.decayFactor, float64(relativeDecayTime+1)))/calculator.decayRate - 1

	gainedMana := float64(movedCoins) * calculator.manaScaleFactor * erosionFactor

	result = uint64(math.Round(gainedMana))
	roundingError = gainedMana - float64(result)

	return
}

func (calculator *Calculator) ManaOfTransferContinuous1(movedCoins uint64, heldTime uint64) (result uint64, roundingError float64) {
	gainsInConsecutiveIntervals := (float64(movedCoins) / calculator.decayRate) * (1 - math.Pow(math.E, -calculator.decayRate*float64(heldTime)))

	result = uint64(gainsInConsecutiveIntervals)
	roundingError = gainsInConsecutiveIntervals - float64(result)

	return
}

func (calculator *Calculator) ErodedManaContinuous(mana uint64, erosionTime uint64) (result uint64, roundingError float64) {
	if erosionTime == 0 {
		result = mana
		roundingError = 0

		return
	}

	growthFactor := math.Log(1-calculator.decayRate) / float64(calculator.decayInterval)
	erodedMana := float64(mana) * math.Pow(math.E, growthFactor*float64(erosionTime))

	result = uint64(erodedMana)
	roundingError = erodedMana - float64(result)

	return
}

func (calculator *Calculator) ErodedManaDiscrete(mana uint64, erosionStartTime uint64, erosionEndTime uint64) (result uint64, lastErosion uint64, roundingError float64) {
	if erosionStartTime > erosionEndTime {
		panic("negative erosion duration")
	}

	erosionIntervals := erosionEndTime/calculator.decayInterval - erosionStartTime/calculator.decayInterval
	erodedValue := math.Pow(float64(1-calculator.decayRate), float64(erosionIntervals)) * float64(mana)

	result = uint64(erodedValue)
	lastErosion = erosionEndTime
	roundingError = erodedValue - float64(result)

	return
}
