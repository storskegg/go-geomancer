package content

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/storskegg/go-geomancer/figure"
)

func TestEmbeddedLocales(t *testing.T) {
	locales := Locales()
	if !slices.Contains(locales, DefaultLocale) {
		t.Fatalf("Locales() = %v, want it to contain %q", locales, DefaultLocale)
	}
	for _, l := range locales {
		s, err := Load(l)
		if err != nil {
			t.Errorf("Load(%q): %v", l, err)
			continue
		}
		if s.Locale != l {
			t.Errorf("%s.json declares locale %q", l, s.Locale)
		}
		for _, f := range figure.All() {
			if got := s.For(f).Figure; got != f {
				t.Errorf("%s: For(%v).Figure = %v", l, f, got)
			}
		}
	}
}

func TestDefault(t *testing.T) {
	if got := Default().Locale; got != DefaultLocale {
		t.Errorf("Default().Locale = %q", got)
	}
}

func TestLoadUnknown(t *testing.T) {
	for _, l := range []string{"xx", "../content"} {
		if _, err := Load(l); !errors.Is(err, ErrUnknownLocale) {
			t.Errorf("Load(%q) error = %v, want ErrUnknownLocale", l, err)
		}
	}
}

func TestParse(t *testing.T) {
	s, err := Parse([]byte(file(`{"figure": "Via", "short": "change", "long": "The road.", "houses": {"1": "travel", "12": "release"}}`)))
	if err != nil {
		t.Fatal(err)
	}
	via := s.For(figure.Via)
	if via.Short != "change" || via.Long != "The road." {
		t.Errorf("Via = %+v", via)
	}
	if via.House(1) != "travel" || via.House(12) != "release" || via.House(2) != "" {
		t.Errorf("Via houses = %q", via.Houses)
	}
	if via.House(0) != "" || via.House(13) != "" {
		t.Error("out-of-range House() not empty")
	}
}

func TestParseInvalid(t *testing.T) {
	tests := map[string]string{
		"not JSON":     `{`,
		"bad figure":   `{"locale": "en", "figures": [{"figure": "Nobody"}]}`,
		"missing":      `{"locale": "en", "figures": [{"figure": "Via"}]}`,
		"duplicate":    file(`{"figure": "Via"}`, `{"figure": "Via"}`),
		"house 13":     file(`{"figure": "Via", "houses": {"13": ""}}`),
		"house letter": file(`{"figure": "Via", "houses": {"one": ""}}`),
		"no figure":    file(`{"short": "unnamed"}`),
	}
	for name, in := range tests {
		if _, err := Parse([]byte(in)); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: error = %v, want ErrInvalid", name, err)
		}
	}
}

func TestEntryJSONRoundTrip(t *testing.T) {
	in := Entry{Figure: figure.Puer, Short: "s", Long: "l"}
	in.Houses[6] = "seventh"
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"houses":{"1":"","2":"",`) || !strings.Contains(string(b), `"7":"seventh"`) {
		t.Errorf("Marshal = %s", b)
	}
	var out Entry
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out != in {
		t.Errorf("round trip = %+v, want %+v", out, in)
	}
}

// file returns a content file containing extra, plus an empty entry for
// every figure not named in extra.
func file(extra ...string) string {
	entries := slices.Clone(extra)
	all := strings.Join(extra, "")
	for _, f := range figure.All() {
		if !strings.Contains(all, `"`+f.Name()+`"`) {
			entries = append(entries, `{"figure": "`+f.Name()+`"}`)
		}
	}
	return `{"locale": "en", "figures": [` + strings.Join(entries, ",") + `]}`
}
