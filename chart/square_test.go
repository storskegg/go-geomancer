package chart

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/system/goldendawn"
	"github.com/storskegg/go-geomancer/system/traditional"
	"github.com/storskegg/go-geomancer/zodiac"
)

// exampleMothers is the hand-worked shield from TestCastWorkedExample.
var exampleMothers = [4]figure.Figure{figure.Amissio, figure.FortunaMinor, figure.Via, figure.Populus}

func exampleShield(t *testing.T) Shield {
	t.Helper()
	s, err := Cast(exampleMothers)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// withSigns returns the traditional placement with the given sign table.
func withSigns(signs map[figure.Figure]zodiac.Sign) system.Mapping {
	m := *traditional.System.(*system.Mapping)
	m.Slug = "test"
	m.Signs = signs
	return m
}

func TestBuildSquarePlacement(t *testing.T) {
	s := exampleShield(t)
	for _, sys := range []system.System{traditional.System, goldendawn.System} {
		t.Run(sys.ID(), func(t *testing.T) {
			sq, err := BuildSquare(s, sys)
			if err != nil {
				t.Fatal(err)
			}
			for i, h := range sq {
				if h.Number != i+1 {
					t.Errorf("square[%d].Number = %d", i, h.Number)
				}
				if want, _ := sys.House(h.Position); want != h.Number {
					t.Errorf("house %d holds %v, which %s places in house %d", h.Number, h.Position, sys.ID(), want)
				}
				if h.Figure != s.At(h.Position) {
					t.Errorf("house %d figure = %v, want %v at %v", h.Number, h.Figure, s.At(h.Position), h.Position)
				}
			}
		})
	}
}

func TestBuildSquareGoldenDawn(t *testing.T) {
	sq, err := BuildSquare(exampleShield(t), goldendawn.System)
	if err != nil {
		t.Fatal(err)
	}
	want := [12]position.Position{
		position.M2, position.D2, position.N2, position.M3,
		position.D3, position.N3, position.M4, position.D4,
		position.N4, position.M1, position.D1, position.N1,
	}
	for i, p := range want {
		if sq[i].Position != p {
			t.Errorf("house %d = %v, want %v", i+1, sq[i].Position, p)
		}
	}
}

func TestBuildSquareSigns(t *testing.T) {
	// M1 (Amissio) is in the 1st House under the traditional placement.
	sq, err := BuildSquare(exampleShield(t), withSigns(map[figure.Figure]zodiac.Sign{figure.Amissio: zodiac.Cancer}))
	if err != nil {
		t.Fatal(err)
	}
	want := []zodiac.Sign{
		zodiac.Cancer, zodiac.Leo, zodiac.Virgo, zodiac.Libra, zodiac.Scorpio, zodiac.Sagittarius,
		zodiac.Capricorn, zodiac.Aquarius, zodiac.Pisces, zodiac.Aries, zodiac.Taurus, zodiac.Gemini,
	}
	for i, h := range sq {
		if h.Sign == nil || *h.Sign != want[i] {
			t.Errorf("house %d sign = %v, want %v", h.Number, h.Sign, want[i])
		}
	}
}

func TestBuildSquareUnmappedFirstHouse(t *testing.T) {
	// Via is in the square (M3), but the 1st House figure has no sign.
	sq, err := BuildSquare(exampleShield(t), withSigns(map[figure.Figure]zodiac.Sign{figure.Via: zodiac.Leo}))
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range sq {
		if h.Sign != nil {
			t.Errorf("house %d sign = %v, want none", h.Number, *h.Sign)
		}
	}
}

func TestBuildSquareInvalidSystem(t *testing.T) {
	m := withSigns(nil)
	m.Houses[position.N4] = 1
	if _, err := BuildSquare(exampleShield(t), m); !errors.Is(err, system.ErrInvalid) {
		t.Errorf("BuildSquare error = %v, want system.ErrInvalid", err)
	}
}

func TestHouseJSON(t *testing.T) {
	sign := zodiac.Cancer
	tests := []struct {
		house House
		want  string
	}{
		{House{1, position.M1, figure.Amissio, &sign}, `{"number":1,"position":"M1","figure":"Amissio","sign":"Cancer"}`},
		{House{2, position.M2, figure.Via, nil}, `{"number":2,"position":"M2","figure":"Via"}`},
	}
	for _, tt := range tests {
		b, err := json.Marshal(tt.house)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != tt.want {
			t.Errorf("Marshal = %s, want %s", b, tt.want)
		}
	}
}

func TestNewReading(t *testing.T) {
	r, err := NewReading(exampleMothers, goldendawn.System)
	if err != nil {
		t.Fatal(err)
	}
	if r.System != "goldendawn" {
		t.Errorf("System = %q", r.System)
	}
	if r.Shield != exampleShield(t) {
		t.Error("Shield differs from Cast")
	}
	if r.Square[0].Position != position.M2 {
		t.Errorf("1st House = %v, want M2", r.Square[0].Position)
	}

	if _, err := NewReading([4]figure.Figure{16}, goldendawn.System); !errors.Is(err, ErrInvalidMother) {
		t.Errorf("NewReading error = %v, want ErrInvalidMother", err)
	}
}
