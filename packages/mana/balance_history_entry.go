package mana

type BalanceHistoryEntry struct {
	transfer                 *Transfer
	balance                  uint64
	accumulatedRoundingError float64
}
