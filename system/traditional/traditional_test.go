package traditional

import (
	"testing"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/zodiac"
)

func TestRegistered(t *testing.T) {
	if s, ok := system.Get("traditional"); !ok || s != System {
		t.Fatalf("Get(traditional) = %v, %v", s, ok)
	}
	if err := system.Validate(System); err != nil {
		t.Error(err)
	}
	if got := System.Name(); got != "Traditional" {
		t.Errorf("Name() = %q", got)
	}
}

func TestSegmentElements(t *testing.T) {
	want := map[figure.Segment]zodiac.Element{
		figure.Head: zodiac.Fire,
		figure.Neck: zodiac.Air,
		figure.Body: zodiac.Water,
		figure.Feet: zodiac.Earth,
	}
	for seg, e := range want {
		if got := System.SegmentElement(seg); got != e {
			t.Errorf("SegmentElement(%v) = %v, want %v", seg, got, e)
		}
	}
}

func TestHouses(t *testing.T) {
	for i, p := range position.Principals() {
		if h, ok := System.House(p); !ok || h != i+1 {
			t.Errorf("House(%v) = %d, %v, want %d", p, h, ok, i+1)
		}
	}
}

// Every figure has an element, and every figure but the lunar nodes has a
// sign.
func TestFigureTablesComplete(t *testing.T) {
	for _, f := range figure.All() {
		_, hasSign := System.Sign(f)
		if node := f == figure.CaputDraconis || f == figure.CaudaDraconis; hasSign == node {
			t.Errorf("%v: has sign = %v, want %v", f, hasSign, !node)
		}
		if _, ok := System.Element(f); !ok {
			t.Errorf("%v has no element", f)
		}
	}
}
