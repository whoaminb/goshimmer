package mana

import (
	"testing"
)

func BenchmarkDirectReceiver(b *testing.B) {
	input := &Input{
		coinAmount:   10,
		receivedTime: 0,
	}

	for i := 0; i < b.N; i++ {
		input.GetCoinAmount()
	}
}

func BenchmarkManualReceiver(b *testing.B) {
	input := &Input{
		coinAmount:   10,
		receivedTime: 0,
	}

	for i := 0; i < b.N; i++ {
		input_getCoinAmount(input)
	}
}
