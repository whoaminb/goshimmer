package bitutils

import (
	"fmt"
	"testing"
)

func BenchmarkOffset_Inc(b *testing.B) {
	var x Offset

	for i := 0; i < b.N; i++ {
		x.Inc(1)
	}
}

func Benchmark_Inc(b *testing.B) {
	x := 0

	for i := 0; i < b.N; i++ {
		x++
	}
}

func TestOffset_Inc(t *testing.T) {
	q := 3

	j := &q

	*j = 12

	fmt.Println(q)

	var x Offset

	x.Inc(4)

	fmt.Println(x)
}
