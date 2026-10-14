package figure

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestFigures(t *testing.T) {
	tests := []struct {
		fig     Figure
		name    string
		pattern string
		dots    int
	}{
		{Via, "Via", "1111", 4},
		{Populus, "Populus", "2222", 8},
		{Albus, "Albus", "2212", 7},
		{Conjunctio, "Conjunctio", "2112", 6},
		{Puella, "Puella", "1211", 5},
		{Amissio, "Amissio", "1212", 6},
		{FortunaMajor, "Fortuna Major", "2211", 6},
		{FortunaMinor, "Fortuna Minor", "1122", 6},
		{Puer, "Puer", "1121", 5},
		{Rubeus, "Rubeus", "2122", 7},
		{Acquisitio, "Acquisitio", "2121", 6},
		{Laetitia, "Laetitia", "1222", 7},
		{Carcer, "Carcer", "1221", 6},
		{Tristitia, "Tristitia", "2221", 7},
		{CaputDraconis, "Caput Draconis", "2111", 5},
		{CaudaDraconis, "Cauda Draconis", "1112", 5},
	}
	if len(tests) != 16 {
		t.Fatalf("expected 16 cases, got %d", len(tests))
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fig.Name(); got != tt.name {
				t.Errorf("Name() = %q, want %q", got, tt.name)
			}
			if got := tt.fig.Pattern(); got != tt.pattern {
				t.Errorf("Pattern() = %q, want %q", got, tt.pattern)
			}
			if got := tt.fig.Dots(); got != tt.dots {
				t.Errorf("Dots() = %d, want %d", got, tt.dots)
			}
			odd := tt.fig.OddDots()
			if even := tt.fig.EvenDots(); odd+even != tt.dots || even%2 != 0 {
				t.Errorf("OddDots() = %d, EvenDots() = %d, want sum %d with even EvenDots", odd, even, tt.dots)
			}
		})
	}
}

func TestAllIsComplete(t *testing.T) {
	seen := map[Figure]bool{}
	names := map[string]bool{}
	for _, f := range All() {
		if !f.Valid() {
			t.Errorf("All() contains invalid figure %d", f)
		}
		seen[f] = true
		names[f.Name()] = true
	}
	if len(seen) != 16 || len(names) != 16 {
		t.Errorf("All() has %d distinct figures and %d distinct names, want 16 each", len(seen), len(names))
	}
}

func TestSegmentsRoundTrip(t *testing.T) {
	for _, f := range All() {
		if got := FromSegments(f.Segments()); got != f {
			t.Errorf("FromSegments(%v.Segments()) = %v", f, got)
		}
	}
	if got := New(Odd, Even, Odd, Even); got != Amissio {
		t.Errorf("New(odd, even, odd, even) = %v, want Amissio", got)
	}
	if got := Laetitia.Segment(Head); got != Odd {
		t.Errorf("Laetitia head = %v, want odd", got)
	}
	if got := Laetitia.Segment(Feet); got != Even {
		t.Errorf("Laetitia feet = %v, want even", got)
	}
}

func TestParityAdd(t *testing.T) {
	tests := []struct{ a, b, want Parity }{
		{Odd, Even, Odd},
		{Even, Odd, Odd},
		{Odd, Odd, Even},
		{Even, Even, Even},
	}
	for _, tt := range tests {
		if got := tt.a.Add(tt.b); got != tt.want {
			t.Errorf("%v + %v = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

// The worked example from the project brief.
func TestAddExample(t *testing.T) {
	m1 := mustParse(t, "1212")
	m2 := mustParse(t, "1122")
	if got := m1.Add(m2).Pattern(); got != "2112" {
		t.Errorf("1212 + 1122 = %s, want 2112", got)
	}
}

func TestAddProperties(t *testing.T) {
	for _, a := range All() {
		if got := a.Add(Populus); got != a {
			t.Errorf("%v + Populus = %v, want identity", a, got)
		}
		if got := a.Add(a); got != Populus {
			t.Errorf("%v + itself = %v, want Populus", a, got)
		}
		for _, b := range All() {
			if a.Add(b) != b.Add(a) {
				t.Errorf("%v + %v is not commutative", a, b)
			}
			for _, s := range Segments() {
				if got, want := a.Add(b).Segment(s), a.Segment(s).Add(b.Segment(s)); got != want {
					t.Errorf("(%v + %v).%v = %v, want %v", a, b, s, got, want)
				}
			}
			for _, c := range All() {
				if a.Add(b).Add(c) != a.Add(b.Add(c)) {
					t.Errorf("(%v + %v) + %v is not associative", a, b, c)
				}
			}
		}
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		in   string
		want Figure
	}{
		{"1212", Amissio},
		{" 2222 ", Populus},
		{"odd,even,odd,even", Amissio},
		{"Odd, Odd, Even, Even", FortunaMinor},
		{"Fortuna Major", FortunaMajor},
		{"fortunamajor", FortunaMajor},
		{"fortuna_major", FortunaMajor},
		{"Fortuna Maior", FortunaMajor},
		{"CAPUT DRACONIS", CaputDraconis},
		{"Coniunctio", Conjunctio},
		{"Letitia", Laetitia},
	}
	for _, tt := range tests {
		got, err := Parse(tt.in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Parse(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestParseRoundTrip(t *testing.T) {
	for _, f := range All() {
		for _, in := range []string{f.Name(), f.Pattern()} {
			if got := mustParse(t, in); got != f {
				t.Errorf("Parse(%q) = %v, want %v", in, got, f)
			}
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, in := range []string{"", "1213", "121", "12121", "odd,even,odd", "odd,even,odd,maybe", "Nobody"} {
		if _, err := Parse(in); !errors.Is(err, ErrInvalid) {
			t.Errorf("Parse(%q) error = %v, want ErrInvalid", in, err)
		}
	}
}

func TestJSON(t *testing.T) {
	b, err := json.Marshal([]Figure{FortunaMajor, Via})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `["Fortuna Major","Via"]`; got != want {
		t.Errorf("Marshal = %s, want %s", got, want)
	}

	var figs []Figure
	if err := json.Unmarshal([]byte(`["Fortuna Major","1212"]`), &figs); err != nil {
		t.Fatal(err)
	}
	if len(figs) != 2 || figs[0] != FortunaMajor || figs[1] != Amissio {
		t.Errorf("Unmarshal = %v", figs)
	}

	if _, err := json.Marshal(Figure(16)); !errors.Is(err, ErrInvalid) {
		t.Errorf("Marshal(Figure(16)) error = %v, want ErrInvalid", err)
	}
}

func mustParse(t *testing.T, s string) Figure {
	t.Helper()
	f, err := Parse(s)
	if err != nil {
		t.Fatalf("Parse(%q): %v", s, err)
	}
	return f
}

func TestInvalidFigure(t *testing.T) {
	f := Figure(16)
	if f.Valid() {
		t.Error("Figure(16) reported valid")
	}
	if got := f.Name(); got != "Figure(16)" {
		t.Errorf("Name() = %q", got)
	}
	if got := f.String(); got != "Figure(16)" {
		t.Errorf("String() = %q", got)
	}
	for _, f := range All() {
		if f.String() != f.Name() {
			t.Errorf("%v: String() = %q, Name() = %q", f, f.String(), f.Name())
		}
	}
}

func TestAddStaysValid(t *testing.T) {
	// High bits on an out-of-range operand are masked off.
	if got := Via.Add(Figure(0b1_0000)); !got.Valid() || got != Via {
		t.Errorf("Via + Figure(16) = %d, want Via", got)
	}
}

func TestUnmarshalTextInvalid(t *testing.T) {
	f := Puer
	if err := f.UnmarshalText([]byte("Nobody")); !errors.Is(err, ErrInvalid) {
		t.Errorf("UnmarshalText(Nobody) error = %v, want ErrInvalid", err)
	}
	if f != Puer {
		t.Errorf("failed UnmarshalText changed the figure to %v", f)
	}
	var figs []Figure
	if err := json.Unmarshal([]byte(`["Via", "1213"]`), &figs); !errors.Is(err, ErrInvalid) {
		t.Errorf("json.Unmarshal error = %v, want ErrInvalid", err)
	}
}

func TestOddEvenDotsByCount(t *testing.T) {
	// A figure's dots depend only on how many segments are odd.
	for _, f := range All() {
		odd := f.OddDots()
		if want := 8 - odd; f.Dots() != want {
			t.Errorf("%v: Dots() = %d, want 8 - %d odd = %d", f, f.Dots(), odd, want)
		}
	}
}
