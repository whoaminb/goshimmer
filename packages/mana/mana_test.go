package mana

import (
	"fmt"
	"math"
	"testing"

	"github.com/magiconair/properties/assert"
)

func BenchmarkCalculator_ManaOfTransferContinuous(b *testing.B) {
	calculator := NewCalculator(10, 0.1, 1)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		calculator.ManaOfTransferContinuous(10000000, 100000000000)
	}
}

func TestBalance_Erode(t *testing.T) {
	calculator := NewCalculator(50, 0.1, 9)
	balance := NewBalance(calculator)

	calculator1 := NewCalculator(5000, 0.2, 1)

	fmt.Println("===")
	fmt.Println(calculator1.ManaOfTransferContinuous(1000, 5000000))
	fmt.Println(calculator.ManaOfTransferDiscrete(1000, 0, 50))

	balance.AddTransfer(500, 0, 10)

	assert.Equal(t, balance.GetValue(), uint64(450))

	balance.Erode(20)

	assert.Equal(t, balance.GetValue(), uint64(405))

	balance.Erode(40)

	assert.Equal(t, balance.GetValue(), uint64(328))
}

func calcManaContinuous(calculator *Calculator, coins uint64, timeHeld uint64) (result uint64, roundingError float64) {
	scaleFactor := 1 - math.Pow(1-calculator.decayRate, float64(timeHeld)/float64(calculator.decayInterval))
	fmt.Println(scaleFactor)
	erodedGains := float64(coins) * float64(calculator.decayInterval) * scaleFactor

	result = uint64(erodedGains)
	roundingError = erodedGains - float64(result)

	return
}

func TestCalculator_ManaOfTransfer(t *testing.T) {
	manaCalculator := NewCalculator(10, 0.1, 10)

	var mana, lastErosion uint64
	var roundingError float64

	mana, lastErosion, _ = manaCalculator.ManaOfTransferDiscrete(49, 0, 0)
	assert.Equal(t, mana, uint64(0))
	assert.Equal(t, lastErosion, uint64(0))

	fmt.Println(calcManaContinuous(manaCalculator, 50, 10))
	fmt.Println(manaCalculator.ManaOfTransferDiscrete(50, 0, 10))
	fmt.Println(calcManaContinuous(manaCalculator, 50, 20))
	fmt.Println(manaCalculator.ManaOfTransferDiscrete(50, 0, 20))

	mana, lastErosion, _ = manaCalculator.ManaOfTransferDiscrete(49, 0, 1)
	assert.Equal(t, mana, uint64(4))
	assert.Equal(t, lastErosion, uint64(0))

	mana, lastErosion, _ = manaCalculator.ManaOfTransferDiscrete(49, 0, 10)
	assert.Equal(t, mana, uint64(36))
	assert.Equal(t, lastErosion, uint64(10))

	mana, lastErosion, _ = manaCalculator.ManaOfTransferDiscrete(50, 0, 31)
	assert.Equal(t, mana, uint64(101))
	assert.Equal(t, lastErosion, uint64(30))

	fmt.Println(roundingError)
}
