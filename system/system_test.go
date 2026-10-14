package system

import (
	"errors"
	"strings"
	"testing"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/zodiac"
)

func validMapping(id string) Mapping {
	return Mapping{
		Slug:            id,
		Title:           "Test",
		SegmentElements: [4]zodiac.Element{zodiac.Fire, zodiac.Air, zodiac.Water, zodiac.Earth},
		Houses:          [12]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		Signs:           map[figure.Figure]zodiac.Sign{figure.Via: zodiac.Cancer},
		Elements:        map[figure.Figure]zodiac.Element{figure.Via: zodiac.Water},
	}
}

func TestValidateValid(t *testing.T) {
	if err := Validate(validMapping("test")); err != nil {
		t.Error(err)
	}
	empty := validMapping("test")
	empty.Signs, empty.Elements = nil, nil
	if err := Validate(empty); err != nil {
		t.Errorf("empty sign and element tables: %v", err)
	}
}

func TestValidateInvalid(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Mapping)
		want   string
	}{
		{"empty ID", func(m *Mapping) { m.Slug = "" }, "ID"},
		{"uppercase ID", func(m *Mapping) { m.Slug = "Test" }, "ID"},
		{"ID with space", func(m *Mapping) { m.Slug = "golden dawn" }, "ID"},
		{"duplicate segment element", func(m *Mapping) { m.SegmentElements[3] = zodiac.Fire }, "both Fire"},
		{"invalid segment element", func(m *Mapping) { m.SegmentElements[0] = zodiac.Element(9) }, "invalid element"},
		{"unset house", func(m *Mapping) { m.Houses[position.N4] = 0 }, "house 0"},
		{"house out of range", func(m *Mapping) { m.Houses[position.N4] = 13 }, "house 13"},
		{"duplicate house", func(m *Mapping) { m.Houses[position.N4] = 1 }, "both placed in house 1"},
		{"invalid sign", func(m *Mapping) { m.Signs[figure.Puer] = zodiac.Sign(12) }, "invalid sign"},
		{"invalid element", func(m *Mapping) { m.Elements[figure.Puer] = zodiac.Element(4) }, "invalid element"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validMapping("test")
			tt.mutate(&m)
			err := Validate(m)
			if !errors.Is(err, ErrInvalid) {
				t.Fatalf("Validate error = %v, want ErrInvalid", err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Validate error = %q, want it to mention %q", err, tt.want)
			}
		})
	}
}

// nonPrincipalHouse places the Judge in a house, which Mapping can't express.
type nonPrincipalHouse struct{ Mapping }

func (n nonPrincipalHouse) House(p position.Position) (int, bool) {
	if p == position.J {
		return 1, true
	}
	return n.Mapping.House(p)
}

func TestValidateNonPrincipalHouse(t *testing.T) {
	err := Validate(nonPrincipalHouse{validMapping("test")})
	if !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), "not a principal position") {
		t.Errorf("Validate error = %v", err)
	}
}

func TestMappingLookups(t *testing.T) {
	m := validMapping("test")
	if h, ok := m.House(position.D3); !ok || h != 7 {
		t.Errorf("House(D3) = %d, %v, want 7, true", h, ok)
	}
	if _, ok := m.House(position.WR); ok {
		t.Error("House(WR) ok, want false")
	}
	if s, ok := m.Sign(figure.Via); !ok || s != zodiac.Cancer {
		t.Errorf("Sign(Via) = %v, %v", s, ok)
	}
	if _, ok := m.Sign(figure.Puer); ok {
		t.Error("Sign(Puer) ok, want unmapped")
	}
	if _, ok := (Mapping{}).Element(figure.Puer); ok {
		t.Error("Element on nil table ok, want unmapped")
	}
	if got := m.SegmentElement(figure.Body); got != zodiac.Water {
		t.Errorf("SegmentElement(Body) = %v", got)
	}
}

func TestRegistry(t *testing.T) {
	Register(validMapping("test-b"))
	Register(validMapping("test-a"))

	if s, ok := Get(" TEST-A "); !ok || s.ID() != "test-a" {
		t.Errorf("Get(TEST-A) = %v, %v", s, ok)
	}
	if _, ok := Get("missing"); ok {
		t.Error("Get(missing) ok")
	}

	var ids []string
	for _, s := range List() {
		ids = append(ids, s.ID())
	}
	if got := strings.Join(ids, ","); got != "test-a,test-b" {
		t.Errorf("List() IDs = %s", got)
	}

	assertPanics(t, "duplicate", func() { Register(validMapping("test-a")) })
	assertPanics(t, "invalid", func() { Register(validMapping("")) })
}

func assertPanics(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		t.Helper()
		if recover() == nil {
			t.Errorf("%s: did not panic", name)
		}
	}()
	f()
}
