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
		figure.Puer: zodiac.Aries, figure.Acquisitio: zodiac.Aries,
		figure.FortunaMinor: zodiac.Taurus, figure.Albus: zodiac.Gemini,
		figure.Via: zodiac.Leo, figure.Conjunctio: zodiac.Virgo, figure.CaputDraconis: zodiac.Virgo,
		figure.Amissio: zodiac.Libra, figure.Puella: zodiac.Libra, figure.Rubeus: zodiac.Scorpio,
		figure.CaudaDraconis: zodiac.Sagittarius, figure.Populus: zodiac.Capricorn,
		figure.FortunaMajor: zodiac.Aquarius, figure.Tristitia: zodiac.Aquarius,
		figure.Carcer: zodiac.Pisces, figure.Laetitia: zodiac.Pisces,
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

// The Draconis figures have elements and lunar nodes, and in this system
// signs too (see TestSigns).
func TestDraconis(t *testing.T) {
	if e, _ := System.Element(figure.CaputDraconis); e != zodiac.Earth {
		t.Errorf("Caput Draconis element = %v, want Earth", e)
	}
	if e, _ := System.Element(figure.CaudaDraconis); e != zodiac.Fire {
		t.Errorf("Cauda Draconis element = %v, want Fire", e)
	}
}
