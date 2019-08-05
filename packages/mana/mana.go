package mana

func CoinTimeDestroyed(transferredValue uint64, parkedTime uint64) uint64 {
	return transferredValue * parkedTime
}

const (
	DECAY_INTERVAL = 10
	DECAY_RATE     = 0.01
	COINS_PER_MANA = 10
)
