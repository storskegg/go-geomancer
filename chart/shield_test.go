package chart

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/position"
)

// A shield worked out by hand from Mothers 1212, 1122, 1111, 2222.
func TestCastWorkedExample(t *testing.T) {
	s, err := Cast([4]figure.Figure{figure.Amissio, figure.FortunaMinor, figure.Via, figure.Populus})
	if err != nil {
		t.Fatal(err)
	}
	want := map[position.Position]figure.Figure{
		position.M1: figure.Amissio,       // 1212
		position.M2: figure.FortunaMinor,  // 1122
		position.M3: figure.Via,           // 1111
		position.M4: figure.Populus,       // 2222
		position.D1: figure.CaudaDraconis, // 1112 (heads)
		position.D2: figure.Conjunctio,    // 2112 (necks)
		position.D3: figure.Amissio,       // 1212 (bodies)
		position.D4: figure.Albus,         // 2212 (feet)
		position.N1: figure.Conjunctio,    // 1212 + 1122 = 2112
		position.N2: figure.Via,           // 1111 + 2222 = 1111
		position.N3: figure.Laetitia,      // 1112 + 2112 = 1222
		position.N4: figure.Laetitia,      // 1212 + 2212 = 1222
		position.WR: figure.Carcer,        // 2112 + 1111 = 1221
		position.WL: figure.Populus,       // 1222 + 1222 = 2222
		position.J:  figure.Carcer,        // 1221 + 2222 = 1221
		position.JM: figure.FortunaMajor,  // 1221 + 1212 = 2211
	}
	for _, p := range position.All() {
		if got := s.At(p); got != want[p] {
			t.Errorf("%v = %v (%s), want %v (%s)", p, got, got.Pattern(), want[p], want[p].Pattern())
		}
	}
}

// TestCastExhaustive checks every possible set of Mothers against the
// derivation rules and the known invariants of the Shield.
func TestCastExhaustive(t *testing.T) {
	all := figure.All()
	for _, m1 := range all {
		for _, m2 := range all {
			for _, m3 := range all {
				for _, m4 := range all {
					mothers := [4]figure.Figure{m1, m2, m3, m4}
					s, err := Cast(mothers)
					if err != nil {
						t.Fatal(err)
					}
					if !checkShield(t, s) {
						t.Fatalf("mothers %v", mothers)
					}
				}
			}
		}
	}
}

func checkShield(t *testing.T, s Shield) bool {
	t.Helper()
	ok := true
	fail := func(format string, args ...any) {
		t.Helper()
		t.Errorf(format, args...)
		ok = false
	}

	for i, seg := range figure.Segments() {
		for j, m := range s.Mothers {
			if got, want := s.Daughters[i].Segment(figure.Segment(j)), m.Segment(seg); got != want {
				fail("D%d.%v = %v, want M%d.%v = %v", i+1, figure.Segment(j), got, j+1, seg, want)
			}
		}
	}

	at := s.At
	sums := []struct{ p, a, b position.Position }{
		{position.N1, position.M1, position.M2},
		{position.N2, position.M3, position.M4},
		{position.N3, position.D1, position.D2},
		{position.N4, position.D3, position.D4},
		{position.WR, position.N1, position.N2},
		{position.WL, position.N3, position.N4},
		{position.J, position.WR, position.WL},
		{position.JM, position.J, position.M1},
	}
	for _, sum := range sums {
		if got, want := at(sum.p), at(sum.a).Add(at(sum.b)); got != want {
			fail("%v = %v, want %v + %v = %v", sum.p, got, sum.a, sum.b, want)
		}
	}

	// The Right Witness sums the Mothers; the Left Witness sums the Daughters.
	if got, want := s.RightWitness, s.Mothers[0].Add(s.Mothers[1]).Add(s.Mothers[2]).Add(s.Mothers[3]); got != want {
		fail("WR = %v, want sum of Mothers %v", got, want)
	}
	if got, want := s.LeftWitness, s.Daughters[0].Add(s.Daughters[1]).Add(s.Daughters[2]).Add(s.Daughters[3]); got != want {
		fail("WL = %v, want sum of Daughters %v", got, want)
	}

	// The Judge always has an even number of dots.
	if s.Judge.Dots()%2 != 0 {
		fail("Judge %v has %d dots, want even", s.Judge, s.Judge.Dots())
	}
	return ok
}

func TestCastInvalidMother(t *testing.T) {
	_, err := Cast([4]figure.Figure{figure.Via, figure.Figure(16), figure.Via, figure.Via})
	if !errors.Is(err, ErrInvalidMother) {
		t.Errorf("Cast error = %v, want ErrInvalidMother", err)
	}
}

func TestAtPanicsOnInvalidPosition(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("At(Position(16)) did not panic")
		}
	}()
	Shield{}.At(position.Position(16))
}

func TestShieldJSON(t *testing.T) {
	s, err := Cast([4]figure.Figure{figure.Amissio, figure.FortunaMinor, figure.Via, figure.Populus})
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"mothers":["Amissio","Fortuna Minor","Via","Populus"],` +
		`"daughters":["Cauda Draconis","Conjunctio","Amissio","Albus"],` +
		`"nieces":["Conjunctio","Via","Laetitia","Laetitia"],` +
		`"rightWitness":"Carcer","leftWitness":"Populus","judge":"Carcer","judgement":"Fortuna Major"}`
	if string(b) != want {
		t.Errorf("Marshal =\n%s\nwant\n%s", b, want)
	}

	var back Shield
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back != s {
		t.Errorf("round trip = %+v, want %+v", back, s)
	}
}
