package mana

import (
	"math"
)

func CoinTimeDestroyed(transferredValue uint64, parkedTime uint64) uint64 {
	return transferredValue * parkedTime
}

func ManaOfTransfer(value uint64, receivedTime uint64, spentTime uint64) (result uint64) {
	firstIntervalDuration := DECAY_INTERVAL - receivedTime%DECAY_INTERVAL
	lastIntervalDuration := spentTime % DECAY_INTERVAL
	erosionCount := (spentTime - receivedTime) / DECAY_INTERVAL

	gainsInFirstInterval := CoinTimeDestroyed(value, firstIntervalDuration)
	gainsInLastInterval := CoinTimeDestroyed(value, lastIntervalDuration)
	gainsPerConsecutiveInterval := CoinTimeDestroyed(value, DECAY_INTERVAL)

	scaleFactor := 1 - DECAY_RATE

	erodedGainsOfFirstInterval := uint64(float64(gainsInFirstInterval) * math.Pow(scaleFactor, float64(erosionCount)))

	var erodedGainsOfConsecutiveIntervals uint64
	if erosionCount >= 1 {
		erodedGainsOfConsecutiveIntervals = uint64(float64(gainsPerConsecutiveInterval) * scaleFactor * (1 - math.Pow(scaleFactor, float64(erosionCount-1))) / (1 - scaleFactor))
	}

	result += erodedGainsOfFirstInterval
	result += erodedGainsOfConsecutiveIntervals
	result += gainsInLastInterval

	return
}

const (
	DECAY_INTERVAL = 100
	DECAY_RATE     = 0.5
)
