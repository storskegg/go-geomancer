package position

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestAll(t *testing.T) {
	all := All()
	if len(all) != 16 {
		t.Fatalf("len(All()) = %d, want 16", len(all))
	}
	for i, p := range all {
		if !p.Valid() {
			t.Errorf("All()[%d] = %v is invalid", i, p)
		}
	}
	if Position(16).Valid() {
		t.Error("Position(16) reported valid")
	}
}

func TestPrincipals(t *testing.T) {
	principals := map[Position]bool{}
	for _, p := range Principals() {
		principals[p] = true
	}
	for _, p := range All() {
		if p.Principal() != principals[p] {
			t.Errorf("%v.Principal() = %v, want %v", p, p.Principal(), principals[p])
		}
	}
	if len(principals) != 12 {
		t.Errorf("Principals() has %d distinct positions, want 12", len(principals))
	}
}

func TestParse(t *testing.T) {
	for _, p := range All() {
		if got, err := Parse(p.String()); err != nil || got != p {
			t.Errorf("Parse(%q) = %v, %v", p, got, err)
		}
	}
	if got, err := Parse(" jm "); err != nil || got != JM {
		t.Errorf("Parse(jm) = %v, %v", got, err)
	}
	for _, in := range []string{"", "M5", "Judge"} {
		if _, err := Parse(in); !errors.Is(err, ErrUnknown) {
			t.Errorf("Parse(%q) error = %v, want ErrUnknown", in, err)
		}
	}
}

func TestJSON(t *testing.T) {
	b, err := json.Marshal(map[Position]int{WR: 1})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `{"WR":1}`; got != want {
		t.Errorf("Marshal = %s, want %s", got, want)
	}
	if _, err := json.Marshal(Position(16)); !errors.Is(err, ErrUnknown) {
		t.Errorf("Marshal(Position(16)) error = %v, want ErrUnknown", err)
	}
}

func TestString(t *testing.T) {
	want := []string{"M1", "M2", "M3", "M4", "D1", "D2", "D3", "D4", "N1", "N2", "N3", "N4", "WR", "WL", "J", "JM"}
	for i, p := range All() {
		if got := p.String(); got != want[i] {
			t.Errorf("Position(%d).String() = %q, want %q", i, got, want[i])
		}
	}
	if got := Position(16).String(); got != "Position(16)" {
		t.Errorf("Position(16).String() = %q", got)
	}
}

func TestPrincipalsOrder(t *testing.T) {
	want := [12]Position{M1, M2, M3, M4, D1, D2, D3, D4, N1, N2, N3, N4}
	if got := Principals(); got != want {
		t.Errorf("Principals() = %v", got)
	}
}

func TestUnmarshalText(t *testing.T) {
	var ps []Position
	if err := json.Unmarshal([]byte(`["M1", "wr", "JM"]`), &ps); err != nil {
		t.Fatal(err)
	}
	if len(ps) != 3 || ps[0] != M1 || ps[1] != WR || ps[2] != JM {
		t.Errorf("Unmarshal = %v", ps)
	}

	p := J
	if err := p.UnmarshalText([]byte("X1")); !errors.Is(err, ErrUnknown) || p != J {
		t.Errorf("UnmarshalText(X1) = %v, left %v", err, p)
	}
}
