package zodiac

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestLists(t *testing.T) {
	if got := len(Elements()); got != 4 {
		t.Errorf("len(Elements()) = %d, want 4", got)
	}
	if got := len(Planets()); got != 9 {
		t.Errorf("len(Planets()) = %d, want 9", got)
	}
	if got := len(Signs()); got != 12 {
		t.Errorf("len(Signs()) = %d, want 12", got)
	}
	for i, e := range Elements() {
		if int(e) != i || !e.Valid() {
			t.Errorf("Elements()[%d] = %v", i, e)
		}
	}
	for i, p := range Planets() {
		if int(p) != i || !p.Valid() {
			t.Errorf("Planets()[%d] = %v", i, p)
		}
	}
	for i, s := range Signs() {
		if int(s) != i || !s.Valid() {
			t.Errorf("Signs()[%d] = %v", i, s)
		}
	}
}

func TestInvalidStrings(t *testing.T) {
	tests := []struct{ got, want string }{
		{Element(4).String(), "Element(4)"},
		{Planet(9).String(), "Planet(9)"},
		{Sign(12).String(), "Sign(12)"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("String() = %q, want %q", tt.got, tt.want)
		}
	}
}

func TestMarshalInvalid(t *testing.T) {
	for _, v := range []any{Element(4), Planet(9), Sign(12)} {
		if _, err := json.Marshal(v); !errors.Is(err, ErrUnknown) {
			t.Errorf("Marshal(%v) error = %v, want ErrUnknown", v, err)
		}
	}
}

func TestUnmarshalInvalid(t *testing.T) {
	e, p, s := Water, Mars, Leo
	if err := e.UnmarshalText([]byte("Aether")); !errors.Is(err, ErrUnknown) || e != Water {
		t.Errorf("Element.UnmarshalText(Aether) = %v, left %v", err, e)
	}
	if err := p.UnmarshalText([]byte("Pluto")); !errors.Is(err, ErrUnknown) || p != Mars {
		t.Errorf("Planet.UnmarshalText(Pluto) = %v, left %v", err, p)
	}
	if err := s.UnmarshalText([]byte("Ophiuchus")); !errors.Is(err, ErrUnknown) || s != Leo {
		t.Errorf("Sign.UnmarshalText(Ophiuchus) = %v, left %v", err, s)
	}
}

// Add is consistent with repeated Next, and n and n±12 agree.
func TestSignAddProperties(t *testing.T) {
	for _, s := range Signs() {
		next := s
		for n := range 30 {
			if got := s.Add(n); got != next {
				t.Errorf("%v.Add(%d) = %v, want %v", s, n, got, next)
			}
			if s.Add(n) != s.Add(n-12) || s.Add(-n) != s.Add(12-n) {
				t.Errorf("%v.Add is not periodic at %d", s, n)
			}
			if got := s.Add(n).Add(-n); got != s {
				t.Errorf("%v.Add(%d).Add(-%d) = %v", s, n, n, got)
			}
			next = next.Next()
		}
	}
}
