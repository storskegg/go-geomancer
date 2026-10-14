package chart

import (
	"iter"

	"github.com/storskegg/go-geomancer/figure"
)

// ChartCount is the number of distinct sets of Mothers, 16⁴.
const ChartCount = 1 << 16

// ChartID identifies a set of Mothers by packing their 4-bit figure values,
// M1 in the most significant place: ID = M1<<12 | M2<<8 | M3<<4 | M4. Written
// in hexadecimal, each digit is one Mother.
func ChartID(mothers [4]figure.Figure) int {
	id := 0
	for _, m := range mothers {
		id = id<<4 | int(m&0b1111)
	}
	return id
}

// MothersOf is the inverse of ChartID. It panics if id is out of range.
func MothersOf(id int) [4]figure.Figure {
	if id < 0 || id >= ChartCount {
		panic("chart: chart ID out of range")
	}
	var m [4]figure.Figure
	for i := range m {
		m[i] = figure.Figure(id >> (4 * (3 - i)) & 0b1111)
	}
	return m
}

// AllMothers yields every set of Mothers with its chart ID, in ID order.
func AllMothers() iter.Seq2[int, [4]figure.Figure] {
	return func(yield func(int, [4]figure.Figure) bool) {
		for id := range ChartCount {
			if !yield(id, MothersOf(id)) {
				return
			}
		}
	}
}
