package goors

import (
	"math"
	"math/rand"
	"testing"
)

var result_simple_test int

func BenchmarkSimple(b *testing.B) {
	sum := 0
	size := 85000
	points := make([]Point, size)
	rand.Seed(42)
	for i := 0; i < size; i++ {
		points[i] = Point{float64(rand.Float32() * 100), float64(rand.Float32() * 100)}
	}
	dsAdvanced := NewRangeSearchSimple(points)
	dsAdvanced.Build()
	numberOfQueries := b.N

	b.ResetTimer()
	for i := 0; i < numberOfQueries; i++ {
		x1 := float64(rand.Float32() * 50)
		x2 := float64(rand.Float32() * 50)
		y1 := float64(rand.Float32() * 50)
		y2 := float64(rand.Float32() * 50)

		bottomLeft := Point{math.Min(x1, x2), math.Min(y1, y2)}
		topRight := Point{math.Max(x1, x2), math.Max(y1, y2)}

		sum += len(dsAdvanced.Query(bottomLeft, topRight))
	}

	result_simple_test = sum
}
