package zodiac

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestSignAdd(t *testing.T) {
	tests := []struct {
		from Sign
		n    int
		want Sign
	}{
		{Cancer, 0, Cancer},
		{Cancer, 1, Leo},
		{Cancer, 2, Virgo},
		{Cancer, 11, Gemini},
		{Pisces, 1, Aries},
		{Aries, -1, Pisces},
		{Aries, 24, Aries},
		{Aries, -25, Pisces},
	}
	for _, tt := range tests {
		if got := tt.from.Add(tt.n); got != tt.want {
			t.Errorf("%v.Add(%d) = %v, want %v", tt.from, tt.n, got, tt.want)
		}
	}
}

func TestSignNextCycles(t *testing.T) {
	s := Aries
	for i, want := range Signs() {
		if s != want {
			t.Fatalf("step %d: got %v, want %v", i, s, want)
		}
		s = s.Next()
	}
	if s != Aries {
		t.Errorf("12 steps from Aries reached %v, want Aries", s)
	}
}

func TestParseRoundTrip(t *testing.T) {
	for _, s := range Signs() {
		if got, err := ParseSign(s.String()); err != nil || got != s {
			t.Errorf("ParseSign(%q) = %v, %v", s, got, err)
		}
	}
	for _, p := range Planets() {
		if got, err := ParsePlanet(p.String()); err != nil || got != p {
			t.Errorf("ParsePlanet(%q) = %v, %v", p, got, err)
		}
	}
	for _, e := range Elements() {
		if got, err := ParseElement(e.String()); err != nil || got != e {
			t.Errorf("ParseElement(%q) = %v, %v", e, got, err)
		}
	}
}

func TestParseLenient(t *testing.T) {
	if got, err := ParsePlanet("north_node"); err != nil || got != NorthNode {
		t.Errorf("ParsePlanet(north_node) = %v, %v", got, err)
	}
	if got, err := ParseSign(" sagittarius "); err != nil || got != Sagittarius {
		t.Errorf("ParseSign(sagittarius) = %v, %v", got, err)
	}
}

func TestParseUnknown(t *testing.T) {
	if _, err := ParseSign("Ophiuchus"); !errors.Is(err, ErrUnknown) {
		t.Errorf("ParseSign(Ophiuchus) error = %v, want ErrUnknown", err)
	}
	if _, err := ParseElement(""); !errors.Is(err, ErrUnknown) {
		t.Errorf("ParseElement(\"\") error = %v, want ErrUnknown", err)
	}
}

func TestInvalidValues(t *testing.T) {
	if Sign(12).Valid() || Planet(9).Valid() || Element(4).Valid() {
		t.Error("out-of-range values reported valid")
	}
	if got := Sign(12).String(); got != "Sign(12)" {
		t.Errorf("Sign(12).String() = %q", got)
	}
	if _, err := json.Marshal(Element(4)); !errors.Is(err, ErrUnknown) {
		t.Errorf("Marshal(Element(4)) error = %v, want ErrUnknown", err)
	}
}

func TestJSON(t *testing.T) {
	type payload struct {
		Sign    Sign    `json:"sign"`
		Planet  Planet  `json:"planet"`
		Element Element `json:"element"`
	}
	in := payload{Capricorn, SouthNode, Water}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(b), `{"sign":"Capricorn","planet":"South Node","element":"Water"}`; got != want {
		t.Errorf("Marshal = %s, want %s", got, want)
	}
	var out payload
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Errorf("round trip = %+v, want %+v", out, in)
	}
}
