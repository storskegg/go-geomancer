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

func TestParseInvalidTypes(t *testing.T) {
	tests := map[string]string{
		"houses array":  file(`{"figure": "Via", "houses": ["a"]}`),
		"short number":  file(`{"figure": "Via", "short": 5}`),
		"figures empty": `{"locale": "en", "figures": []}`,
		"house zero":    file(`{"figure": "Via", "houses": {"0": ""}}`),
	}
	for name, in := range tests {
		if _, err := Parse([]byte(in)); err == nil {
			t.Errorf("%s: Parse succeeded", name)
		}
	}
}

func TestHousesJSON(t *testing.T) {
	var h Houses
	h[0] = `a "quoted" line` + "\nand a newline"
	h[11] = "<last> & ünïcode"
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(b) {
		t.Fatalf("Marshal produced invalid JSON: %s", b)
	}
	if !strings.HasPrefix(string(b), `{"1":"a \"quoted\" line\nand a newline","2":""`) {
		t.Errorf("Marshal = %s", b)
	}
	var back Houses
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back != h {
		t.Errorf("round trip = %q, want %q", back, h)
	}
}

// Unmarshalling into a used Houses value replaces it rather than merging.
func TestHousesUnmarshalResets(t *testing.T) {
	h := Houses{"old"}
	if err := json.Unmarshal([]byte(`{"2": "new"}`), &h); err != nil {
		t.Fatal(err)
	}
	if h[0] != "" || h[1] != "new" {
		t.Errorf("houses = %q", h)
	}
}

func TestEmbeddedEntriesUseCanonicalNames(t *testing.T) {
	for _, l := range Locales() {
		b, err := data.ReadFile("data/" + l + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var raw struct {
			Figures []struct {
				Figure string `json:"figure"`
			} `json:"figures"`
		}
		if err := json.Unmarshal(b, &raw); err != nil {
			t.Fatal(err)
		}
		for _, e := range raw.Figures {
			f, err := figure.Parse(e.Figure)
			if err != nil || f.Name() != e.Figure {
				t.Errorf("%s.json: figure %q should be written %q", l, e.Figure, f.Name())
			}
		}
	}
}
