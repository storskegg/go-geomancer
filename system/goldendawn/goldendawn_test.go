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

// Every figure has a sign, an element and a planet.
func TestFigureTablesComplete(t *testing.T) {
	for _, f := range figure.All() {
		if _, ok := System.Sign(f); !ok {
			t.Errorf("%v has no sign", f)
		}
		if _, ok := System.Element(f); !ok {
			t.Errorf("%v has no element", f)
		}
		if _, ok := System.Planet(f); !ok {
			t.Errorf("%v has no planet", f)
		}
	}
}

func TestSigns(t *testing.T) {
	want := map[figure.Figure]zodiac.Sign{
		figure.Puer: zodiac.Aries, figure.Amissio: zodiac.Taurus, figure.Albus: zodiac.Gemini,
		figure.Populus: zodiac.Cancer, figure.Via: zodiac.Cancer,
		figure.FortunaMajor: zodiac.Leo, figure.FortunaMinor: zodiac.Leo,
		figure.Conjunctio: zodiac.Virgo, figure.Puella: zodiac.Libra, figure.Rubeus: zodiac.Scorpio,
		figure.Acquisitio: zodiac.Sagittarius, figure.Carcer: zodiac.Capricorn,
		figure.Tristitia: zodiac.Aquarius, figure.Laetitia: zodiac.Pisces,
		figure.CaputDraconis: zodiac.Virgo, figure.CaudaDraconis: zodiac.Sagittarius,
	}
	if len(want) != 16 {
		t.Fatalf("table has %d figures", len(want))
	}
	for f, sign := range want {
		if got, ok := System.Sign(f); !ok || got != sign {
			t.Errorf("Sign(%v) = %v, %v, want %v", f, got, ok, sign)
		}
	}
}

func TestPlanets(t *testing.T) {
	want := map[figure.Figure]zodiac.Planet{
		figure.Via: zodiac.Moon, figure.Populus: zodiac.Moon,
		figure.Albus: zodiac.Mercury, figure.Conjunctio: zodiac.Mercury,
		figure.Puella: zodiac.Venus, figure.Amissio: zodiac.Venus,
		figure.FortunaMajor: zodiac.Sun, figure.FortunaMinor: zodiac.Sun,
		figure.Puer: zodiac.Mars, figure.Rubeus: zodiac.Mars,
		figure.Acquisitio: zodiac.Jupiter, figure.Laetitia: zodiac.Jupiter,
		figure.Carcer: zodiac.Saturn, figure.Tristitia: zodiac.Saturn,
		figure.CaputDraconis: zodiac.NorthNode, figure.CaudaDraconis: zodiac.SouthNode,
	}
	if len(want) != 16 {
		t.Fatalf("table has %d figures", len(want))
	}
	for f, planet := range want {
		if got, ok := System.Planet(f); !ok || got != planet {
			t.Errorf("Planet(%v) = %v, %v, want %v", f, got, ok, planet)
		}
	}
}

// The Draconis figures have elements and lunar nodes, and signs too (see
// TestSigns).
func TestDraconis(t *testing.T) {
	if e, _ := System.Element(figure.CaputDraconis); e != zodiac.Earth {
		t.Errorf("Caput Draconis element = %v, want Earth", e)
	}
	if e, _ := System.Element(figure.CaudaDraconis); e != zodiac.Fire {
		t.Errorf("Cauda Draconis element = %v, want Fire", e)
	}
}
