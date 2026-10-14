// Package content provides the interpretive text for each figure: a short
// qualitative description, a long description, and a meaning for each of the
// twelve houses.
//
// Text is kept in embedded JSON files under data/, one per locale (data/en.json).
// Empty strings mean the text hasn't been written yet.
package content

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/storskegg/go-geomancer/figure"
)

// DefaultLocale is the locale used by Default.
const DefaultLocale = "en"

// ErrInvalid is returned for a content file that is malformed or doesn't
// describe each figure exactly once.
var ErrInvalid = errors.New("content: invalid content")

// ErrUnknownLocale is returned by Load for a locale with no content file.
var ErrUnknownLocale = errors.New("content: unknown locale")

//go:embed data/*.json
var data embed.FS

// Entry is the text for one figure.
type Entry struct {
	Figure figure.Figure `json:"figure"`
	Short  string        `json:"short"`
	Long   string        `json:"long"`
	Houses Houses        `json:"houses"`
}

// House returns the figure's meaning in house n (1–12), or "" if n is out of
// range.
func (e Entry) House(n int) string {
	if n < 1 || n > len(e.Houses) {
		return ""
	}
	return e.Houses[n-1]
}

func (e *Entry) UnmarshalJSON(b []byte) error {
	type plain Entry
	var raw struct {
		plain
		Figure *figure.Figure `json:"figure"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if raw.Figure == nil {
		return fmt.Errorf("%w: entry has no figure", ErrInvalid)
	}
	*e = Entry(raw.plain)
	e.Figure = *raw.Figure
	return nil
}

// Houses holds a figure's meaning in each house. Index i is House i+1. In
// JSON it is an object keyed "1" through "12".
type Houses [12]string

// MarshalJSON writes the houses in numeric order, which a map would not.
func (h Houses) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, v := range h {
		if i > 0 {
			b.WriteByte(',')
		}
		val, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(&b, "%q:%s", strconv.Itoa(i+1), val)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func (h *Houses) UnmarshalJSON(b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	*h = Houses{}
	for k, v := range m {
		n, err := strconv.Atoi(k)
		if err != nil || n < 1 || n > len(h) {
			return fmt.Errorf("%w: house %q, want 1–12", ErrInvalid, k)
		}
		h[n-1] = v
	}
	return nil
}

// Set is the content for every figure in one locale.
type Set struct {
	Locale  string
	entries [16]Entry
}

// For returns the entry for f.
func (s *Set) For(f figure.Figure) Entry { return s.entries[f&0b1111] }

// Locales returns the locales with embedded content, sorted.
func Locales() []string {
	files, _ := fs.Glob(data, "data/*.json")
	locales := make([]string, 0, len(files))
	for _, f := range files {
		locales = append(locales, strings.TrimSuffix(path.Base(f), ".json"))
	}
	slices.Sort(locales)
	return locales
}

// Load reads the embedded content for a locale.
func Load(locale string) (*Set, error) {
	b, err := data.ReadFile("data/" + locale + ".json")
	if err != nil {
		return nil, fmt.Errorf("%w: %q", ErrUnknownLocale, locale)
	}
	return Parse(b)
}

// Parse decodes a content file and checks it describes each figure exactly
// once.
func Parse(b []byte) (*Set, error) {
	var file struct {
		Locale  string  `json:"locale"`
		Figures []Entry `json:"figures"`
	}
	if err := json.Unmarshal(b, &file); err != nil {
		if errors.Is(err, ErrInvalid) {
			return nil, err
		}
		return nil, fmt.Errorf("%w: %w", ErrInvalid, err)
	}

	s := &Set{Locale: file.Locale}
	var seen [16]bool
	for _, e := range file.Figures {
		if seen[e.Figure] {
			return nil, fmt.Errorf("%w: %v appears more than once", ErrInvalid, e.Figure)
		}
		seen[e.Figure] = true
		s.entries[e.Figure] = e
	}
	for _, f := range figure.All() {
		if !seen[f] {
			return nil, fmt.Errorf("%w: %v is missing", ErrInvalid, f)
		}
	}
	return s, nil
}

var defaultSet = sync.OnceValues(func() (*Set, error) { return Load(DefaultLocale) })

// Default returns the content for DefaultLocale. The embedded file is
// checked by the package tests, so an error here means a broken build.
func Default() *Set {
	s, err := defaultSet()
	if err != nil {
		panic(err)
	}
	return s
}
