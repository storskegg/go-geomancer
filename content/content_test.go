package content

import (
	"encoding/json"
	"errors"
	"fmt"
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
		"bad figure":   fileWith(allSignifications(), []string{`{"figure": "Nobody"}`}),
		"missing":      fileWith(allSignifications(), []string{`{"figure": "Via"}`}),
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
	return fileWith(allSignifications(), entries)
}

// allSignifications returns a valid significations list as JSON.
func allSignifications() string {
	sigs := make([]string, 0, 12)
	for h := 1; h <= 12; h++ {
		sigs = append(sigs, fmt.Sprintf(`{"house": %d, "title": "Title %d", "text": "Text %d"}`, h, h, h))
	}
	return "[" + strings.Join(sigs, ",") + "]"
}

// fileWith assembles a content file from a significations JSON list and
// figure entries.
func fileWith(significations string, entries []string) string {
	return `{"locale": "en", "significations": ` + significations + `, "figures": [` + strings.Join(entries, ",") + `]}`
}

func TestParseInvalidTypes(t *testing.T) {
	tests := map[string]string{
		"houses array":  file(`{"figure": "Via", "houses": ["a"]}`),
		"short number":  file(`{"figure": "Via", "short": 5}`),
		"figures empty": fileWith(allSignifications(), nil),
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

func TestSignifications(t *testing.T) {
	s := Default()
	sigs := s.Significations()
	for i, sig := range sigs {
		if sig.House != i+1 || sig.Title == "" || sig.Text == "" {
			t.Errorf("signification %d = %+v", i+1, sig)
		}
	}
	if sig, ok := s.Signification(7); !ok || sig.Title != "Partnership" {
		t.Errorf("Signification(7) = %+v, %v", sig, ok)
	}
	for _, n := range []int{0, 13} {
		if _, ok := s.Signification(n); ok {
			t.Errorf("Signification(%d) ok", n)
		}
	}
}

func TestFindSignification(t *testing.T) {
	s := Default()
	tests := []struct {
		in   string
		want int // 0 = not found
	}{
		{"7", 7},
		{" 12 ", 12},
		{"partnership", 7},
		{"HONOUR", 10},
		{"13", 0},
		{"marriage", 0}, // matches titles only, not text
		{"", 0},
	}
	for _, tt := range tests {
		sig, ok := s.FindSignification(tt.in)
		if ok != (tt.want != 0) || sig.House != tt.want {
			t.Errorf("FindSignification(%q) = house %d, %v; want %d", tt.in, sig.House, ok, tt.want)
		}
	}
}

func TestParseInvalidSignifications(t *testing.T) {
	var figs []string
	for _, f := range figure.All() {
		figs = append(figs, `{"figure": "`+f.Name()+`"}`)
	}
	without := func(h int) string {
		sigs := make([]string, 0, 12)
		for i := 1; i <= 12; i++ {
			if i != h {
				sigs = append(sigs, fmt.Sprintf(`{"house": %d, "title": "T"}`, i))
			}
		}
		return strings.Join(sigs, ",")
	}
	tests := map[string]string{
		"missing":      "[" + without(5) + "]",
		"duplicate":    "[" + without(0) + `,{"house": 3, "title": "T"}]`,
		"out of range": "[" + without(12) + `,{"house": 13, "title": "T"}]`,
		"no title":     "[" + without(12) + `,{"house": 12, "title": ""}]`,
		"absent":       "[]",
	}
	for name, sigs := range tests {
		if _, err := Parse([]byte(fileWith(sigs, figs))); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: error = %v, want ErrInvalid", name, err)
		}
	}
}

// Every figure has a meaning in every house.
func TestHouseMeaningsWritten(t *testing.T) {
	s := Default()
	for _, f := range figure.All() {
		for h := 1; h <= 12; h++ {
			if s.For(f).House(h) == "" {
				t.Errorf("%v has no meaning in house %d", f, h)
			}
		}
	}
	// Traditionally, Rubeus or Cauda Draconis in House 1 means the chart is
	// not to be judged.
	for _, f := range []figure.Figure{figure.Rubeus, figure.CaudaDraconis} {
		if m := s.For(f).House(1); !strings.Contains(m, "not be judged") {
			t.Errorf("%v in house 1 = %q", f, m)
		}
	}
}

// Each house has terms, weighted 1 to 3, and no person is in two houses.
func TestLexicon(t *testing.T) {
	people := map[string]int{}
	for _, sig := range Default().Significations() {
		if len(sig.Terms) < 20 {
			t.Errorf("house %d has only %d terms", sig.House, len(sig.Terms))
		}
		for term, w := range sig.Terms {
			if w < 1 || w > 3 || strings.TrimSpace(term) != term || strings.ToLower(term) != term {
				t.Errorf("house %d term %q weight %v", sig.House, term, w)
			}
		}
		for _, p := range sig.People {
			if h, dup := people[p]; dup {
				t.Errorf("%q is in houses %d and %d", p, h, sig.House)
			}
			people[p] = sig.House
		}
	}
}
