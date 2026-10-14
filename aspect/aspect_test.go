package aspect

import (
	"encoding/json"
	"errors"
	"testing"
)

// The worked example: with the significator in House 10.
func TestHousesExample(t *testing.T) {
	got, err := Houses(10)
	if err != nil {
		t.Fatal(err)
	}
	want := []House{
		{House: 4, Aspect: Opposition},
		{House: 7, Aspect: Square, Side: Dexter},
		{House: 1, Aspect: Square, Side: Sinister},
		{House: 6, Aspect: Trine, Side: Dexter},
		{House: 2, Aspect: Trine, Side: Sinister},
		{House: 8, Aspect: Sextile, Side: Dexter},
		{House: 12, Aspect: Sextile, Side: Sinister},
	}
	if len(got) != len(want) {
		t.Fatalf("Houses(10) = %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Houses(10)[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// Counting includes the significator: "count four" for the Square reaches
// the house three away.
func TestCountingIncludesSignificator(t *testing.T) {
	counts := map[Aspect]int{Sextile: 3, Square: 4, Trine: 5}
	for a, count := range counts {
		if got := a.Distance(); got != count-1 {
			t.Errorf("%v: distance %d, want count %d less one", a, got, count)
		}
	}
	if Opposition.Distance() != 6 {
		t.Errorf("Opposition distance = %d, want 6", Opposition.Distance())
	}
}

func TestHousesEverySignificator(t *testing.T) {
	for sig := 1; sig <= 12; sig++ {
		houses, err := Houses(sig)
		if err != nil {
			t.Fatal(err)
		}
		seen := map[int]bool{sig: true}
		for _, h := range houses {
			if h.House < 1 || h.House > 12 || seen[h.House] {
				t.Errorf("significator %d: %+v is out of range or repeated", sig, h)
			}
			seen[h.House] = true

			// Dexter counts back through the house numbers (clockwise on
			// the Square), Sinister forward.
			var want int
			switch h.Side {
			case Dexter:
				want = (sig-1-h.Aspect.Distance()+12)%12 + 1
			case Sinister, Neither:
				want = (sig-1+h.Aspect.Distance())%12 + 1
			}
			if h.House != want {
				t.Errorf("significator %d: %v %v is house %d, want %d", sig, h.Side, h.Aspect, h.House, want)
			}
			if (h.Aspect == Opposition) != (h.Side == Neither) {
				t.Errorf("significator %d: %v has side %q", sig, h.Aspect, h.Side)
			}
		}
		if len(houses) != 7 {
			t.Errorf("significator %d: %d houses, want 7", sig, len(houses))
		}
	}
}

// Aspects are symmetric: if A aspects B, B aspects A in the same way, on
// the other side.
func TestHousesSymmetric(t *testing.T) {
	for a := 1; a <= 12; a++ {
		ha, _ := Houses(a)
		for _, h := range ha {
			hb, _ := Houses(h.House)
			found := false
			for _, back := range hb {
				if back.House == a && back.Aspect == h.Aspect {
					found = true
					if h.Side != Neither && back.Side == h.Side {
						t.Errorf("%d→%d %v %v, but back on the same side", a, h.House, h.Side, h.Aspect)
					}
				}
			}
			if !found {
				t.Errorf("%d aspects %d by %v, but not the reverse", a, h.House, h.Aspect)
			}
		}
	}
}

func TestHousesInvalid(t *testing.T) {
	for _, n := range []int{0, 13, -1} {
		if _, err := Houses(n); !errors.Is(err, ErrInvalidHouse) {
			t.Errorf("Houses(%d) error = %v, want ErrInvalidHouse", n, err)
		}
	}
}

func TestNature(t *testing.T) {
	for _, a := range All() {
		if want := a == Opposition || a == Square; a.Stressful() != want {
			t.Errorf("%v.Stressful() = %v", a, a.Stressful())
		}
	}
}

func TestStrings(t *testing.T) {
	want := map[Aspect][2]string{
		Opposition: {"Opposition", "☍"},
		Square:     {"Square", "□"},
		Trine:      {"Trine", "△"},
		Sextile:    {"Sextile", "⚹"},
	}
	for a, w := range want {
		if a.String() != w[0] || a.Symbol() != w[1] {
			t.Errorf("%d: %q %q, want %q %q", a, a.String(), a.Symbol(), w[0], w[1])
		}
	}
	if Aspect(4).Valid() || Aspect(4).String() != "Aspect(4)" {
		t.Error("Aspect(4) is valid or misnamed")
	}
	if Dexter.String() != "Dexter" || Sinister.String() != "Sinister" || Neither.String() != "" || Side(3).String() != "Side(3)" {
		t.Error("side names")
	}
}

func TestJSON(t *testing.T) {
	houses, _ := Houses(10)
	b, err := json.Marshal(houses[:2])
	if err != nil {
		t.Fatal(err)
	}
	want := `[{"house":4,"aspect":"Opposition"},{"house":7,"aspect":"Square","side":"Dexter"}]`
	if string(b) != want {
		t.Errorf("Marshal = %s, want %s", b, want)
	}
	var back []House
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back[0] != houses[0] || back[1] != houses[1] {
		t.Errorf("round trip = %+v", back)
	}

	if _, err := json.Marshal(Aspect(9)); err == nil {
		t.Error("Marshal(Aspect(9)) succeeded")
	}
	var a Aspect
	if err := a.UnmarshalText([]byte("Conjunction")); err == nil {
		t.Error("UnmarshalText(Conjunction) succeeded")
	}
	var s Side
	if err := s.UnmarshalText([]byte("Left")); err == nil {
		t.Error("UnmarshalText(Left) succeeded")
	}
}
