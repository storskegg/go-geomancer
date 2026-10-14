package goldendawn

import (
	"testing"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/zodiac"
)

func TestRegistered(t *testing.T) {
	if s, ok := system.Get("goldendawn"); !ok || s != System {
		t.Fatalf("Get(goldendawn) = %v, %v", s, ok)
	}
	if err := system.Validate(System); err != nil {
		t.Error(err)
	}
	if got := System.Name(); got != "Golden Dawn" {
		t.Errorf("Name() = %q", got)
	}
}

func TestSegmentElements(t *testing.T) {
	want := map[figure.Segment]zodiac.Element{
		figure.Head: zodiac.Fire,
		figure.Neck: zodiac.Water,
		figure.Body: zodiac.Air,
		figure.Feet: zodiac.Earth,
	}
	for seg, e := range want {
		if got := System.SegmentElement(seg); got != e {
			t.Errorf("SegmentElement(%v) = %v, want %v", seg, got, e)
		}
	}
}

func TestHouses(t *testing.T) {
	want := map[position.Position]int{
		position.M1: 10, position.M2: 1, position.M3: 4, position.M4: 7,
		position.D1: 11, position.D2: 2, position.D3: 5, position.D4: 8,
		position.N1: 12, position.N2: 3, position.N3: 6, position.N4: 9,
	}
	for p, house := range want {
		if h, ok := System.House(p); !ok || h != house {
			t.Errorf("House(%v) = %d, %v, want %d", p, h, ok, house)
		}
	}
}
