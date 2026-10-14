package chart

import (
	"testing"

	"github.com/storskegg/go-geomancer/figure"
)

func TestChartID(t *testing.T) {
	tests := []struct {
		mothers [4]figure.Figure
		id      int
	}{
		{[4]figure.Figure{figure.Populus, figure.Populus, figure.Populus, figure.Populus}, 0},
		{[4]figure.Figure{figure.Via, figure.Via, figure.Via, figure.Via}, 0xFFFF},
		// Amissio 0b1010, Fortuna Minor 0b1100, Via 0b1111, Populus 0b0000.
		{[4]figure.Figure{figure.Amissio, figure.FortunaMinor, figure.Via, figure.Populus}, 0xACF0},
	}
	for _, tt := range tests {
		if got := ChartID(tt.mothers); got != tt.id {
			t.Errorf("ChartID(%v) = %#x, want %#x", tt.mothers, got, tt.id)
		}
		if got := MothersOf(tt.id); got != tt.mothers {
			t.Errorf("MothersOf(%#x) = %v, want %v", tt.id, got, tt.mothers)
		}
	}
}

func TestAllMothers(t *testing.T) {
	seen := make(map[[4]figure.Figure]bool, ChartCount)
	next := 0
	for id, m := range AllMothers() {
		if id != next {
			t.Fatalf("got ID %d, want %d", id, next)
		}
		next++
		if ChartID(m) != id {
			t.Fatalf("ChartID(%v) = %d, want %d", m, ChartID(m), id)
		}
		for _, f := range m {
			if !f.Valid() {
				t.Fatalf("ID %d has invalid mother %d", id, f)
			}
		}
		seen[m] = true
	}
	if next != ChartCount || len(seen) != ChartCount {
		t.Errorf("yielded %d IDs, %d distinct sets; want %d", next, len(seen), ChartCount)
	}
}

func TestAllMothersStopsEarly(t *testing.T) {
	n := 0
	for range AllMothers() {
		n++
		if n == 3 {
			break
		}
	}
	if n != 3 {
		t.Errorf("iterated %d times, want 3", n)
	}
}

func TestMothersOfOutOfRange(t *testing.T) {
	for _, id := range []int{-1, ChartCount} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("MothersOf(%d) did not panic", id)
				}
			}()
			MothersOf(id)
		}()
	}
}
