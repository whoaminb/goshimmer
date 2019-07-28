package mana

import "math"

func CoinTimeDestroyed(transferredValue uint64, parkedTime uint64) uint64 {
	return transferredValue * parkedTime
}

func ErodedGains(gainsPerInterval uint64, elapsedIntervals uint64) uint64 {
	return uint64(float64(gainsPerInterval) * DECAY_FACTOR * (1 - math.Pow(DECAY_FACTOR, float64(elapsedIntervals))) / (1 - DECAY_FACTOR))
}

func ManaOfTransfer(value uint64, parkedTime uint64) uint64 {
	fullErosionIntervals := parkedTime / DECAY_INTERVAL

	if fullErosionIntervals >= 1 {
		gainsPerInterval := CoinTimeDestroyed(value, DECAY_INTERVAL)

		return ErodedGains(gainsPerInterval, fullErosionIntervals) + CoinTimeDestroyed(value, (parkedTime-fullErosionIntervals*DECAY_INTERVAL))
	} else {
		return CoinTimeDestroyed(value, parkedTime)
	}
}

const (
	DECAY_INTERVAL = 100
	DECAY_FACTOR   = 0.9
)
