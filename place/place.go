// Package place finds US places (the 50 states and DC) by name, with their
// coordinates, and the time zone at any coordinates. All its data is built
// in: places from GeoNames (CC BY 4.0) and time zone boundaries from
// timezone-boundary-builder via github.com/ringsaturn/tzf (ODbL).
package place

import (
	"bufio"
	"bytes"
	"compress/gzip"
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"unicode"

	"github.com/ringsaturn/tzf"
	"golang.org/x/text/unicode/norm"

	"github.com/storskegg/go-geomancer/chart"
)

// Place is a populated place.
type Place struct {
	Name       string  `json:"name"`             // as written locally, such as "Cañon City"
	State      string  `json:"state"`            // postal abbreviation, such as "CO"
	County     string  `json:"county,omitempty"` // such as "Fremont County"
	Latitude   float64 `json:"latitude"`         // degrees north
	Longitude  float64 `json:"longitude"`        // degrees east; negative is west
	Population int     `json:"population"`
}

// Label names the place "City, ST".
func (p Place) Label() string { return p.Name + ", " + p.State }

// Location returns the place as a chart location.
func (p Place) Location() chart.Location {
	return chart.Location{Name: p.Name, State: p.State, County: p.County, Latitude: p.Latitude, Longitude: p.Longitude}
}

// Circumstances returns the circumstances of a chart cast at t, which must
// be in the place's time zone (see TimeZone and when.Resolve).
func (p Place) Circumstances(t time.Time) chart.Circumstances {
	zone, _ := t.Zone()
	return chart.Circumstances{Time: t, TimeZone: t.Location().String(), Zone: zone, Place: p.Location()}
}

//go:embed data/places.tsv.gz
var placeFile []byte

// Index searches the places.
type Index struct {
	places []Place    // most populous first
	tokens [][]string // per place: name words, then state, state name and county words
	names  []int      // per place: how many of its tokens are the name's
	labels []string   // per place: its label, with the county if another place shares it
	byKey  map[string][]int
}

var loadIndex = sync.OnceValues(func() (*Index, error) { return parse(placeFile) })

// Default returns the built-in index. It's loaded on first use.
func Default() (*Index, error) { return loadIndex() }

func parse(gz []byte) (*Index, error) {
	r, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return nil, fmt.Errorf("place: %w", err)
	}
	ix := &Index{byKey: map[string][]int{}}
	sc := bufio.NewScanner(r)
	count := map[string]int{} // label → places sharing it
	// Counties and states repeat, so split each once.
	cache := map[string][]string{}
	cached := func(s string) []string {
		ws, ok := cache[s]
		if !ok {
			ws = words(s)
			cache[s] = ws
		}
		return ws
	}
	for sc.Scan() {
		f := strings.Split(sc.Text(), "\t")
		if len(f) != 8 {
			return nil, errors.New("place: malformed data")
		}
		lat, err1 := strconv.ParseFloat(f[4], 64)
		lon, err2 := strconv.ParseFloat(f[5], 64)
		pop, err3 := strconv.Atoi(f[6])
		if err1 != nil || err2 != nil || err3 != nil {
			return nil, errors.New("place: malformed data")
		}
		p := Place{Name: f[0], State: f[2], County: f[3], Latitude: lat, Longitude: lon, Population: pop}
		name := words(f[1])
		toks := append(append([]string{}, name...), strings.ToLower(p.State))
		toks = append(toks, cached(stateNames[p.State])...)
		toks = append(toks, cached(p.County)...)
		ix.places = append(ix.places, p)
		ix.tokens = append(ix.tokens, toks)
		ix.names = append(ix.names, len(name))
		count[strings.ToLower(p.Label())]++
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("place: %w", err)
	}
	for i, p := range ix.places {
		label := p.Label()
		if count[strings.ToLower(label)] > 1 && p.County != "" {
			label += " (" + p.County + ")"
		}
		ix.labels = append(ix.labels, label)
		key := strings.Join(words(label), " ")
		ix.byKey[key] = append(ix.byKey[key], i)
	}
	return ix, nil
}

// Len returns how many places there are.
func (ix *Index) Len() int { return len(ix.places) }

// Match is a search result.
type Match struct {
	Place Place
	// Label names it "City, ST", adding the county when another place in
	// the state has the same name: "Springfield, PA (Delaware County)".
	Label string
}

// Search returns up to n places matching the query, most populous first.
// The first word of the query must begin the place's name ("york" finds
// York, PA, not New York City); any others must each begin a word of the
// name, the state's abbreviation or name, or the county: "springfield il",
// "springfield illinois" and "springfield sangamon" all find Springfield,
// IL. Case, accents and punctuation don't matter, and "st", "ft" and "mt"
// match "saint", "fort" and "mount".
func (ix *Index) Search(query string, n int) []Match {
	q := words(query)
	if len(q) == 0 || n <= 0 {
		return nil
	}
	var out []Match
	for i, toks := range ix.tokens {
		if matches(q, toks, ix.names[i]) {
			out = append(out, Match{Place: ix.places[i], Label: ix.labels[i]})
			if len(out) == n {
				break
			}
		}
	}
	return out
}

func matches(q, toks []string, nameLen int) bool {
	if nameLen == 0 || !strings.HasPrefix(toks[0], q[0]) {
		return false
	}
	// The rest, in order among the name's remaining words, or anywhere among
	// the state and county words. A state abbreviation matches only the
	// state: "pa" isn't the start of "Parish" or "Page County".
	next := 1
	for _, w := range q[1:] {
		if stateAbbr[w] {
			if toks[nameLen] != w {
				return false
			}
			continue
		}
		found := false
		for j := next; j < len(toks); j++ {
			if strings.HasPrefix(toks[j], w) {
				found = true
				if j < nameLen {
					next = j + 1
				}
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Find returns the place a label names, as Search labels it, or as people
// write it: "Springfield, IL", "springfield il". Where several places share
// it, it returns the most populous.
func (ix *Index) Find(label string) (Match, bool) {
	is, ok := ix.byKey[strings.Join(words(label), " ")]
	if !ok {
		return Match{}, false
	}
	return Match{Place: ix.places[is[0]], Label: ix.labels[is[0]]}, true
}

// words splits text into lowercase words of letters and digits, without
// accents, expanding "st", "ft" and "mt".
func words(s string) []string {
	var b strings.Builder
	if !isASCII(s) {
		s = norm.NFD.String(s)
	}
	for _, r := range strings.ToLower(s) {
		switch {
		case unicode.Is(unicode.Mn, r):
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '\'' || r == '’':
			// O'Fallon matches "ofallon".
		default:
			b.WriteByte(' ')
		}
	}
	ws := strings.Fields(b.String())
	for i, w := range ws {
		if full, ok := abbreviations[w]; ok {
			ws[i] = full
		}
	}
	return ws
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return false
		}
	}
	return true
}

// stateAbbr holds the lowercase state abbreviations.
var stateAbbr = func() map[string]bool {
	m := map[string]bool{}
	for k := range stateNames {
		m[strings.ToLower(k)] = true
	}
	return m
}()

var abbreviations = map[string]string{"st": "saint", "ste": "sainte", "ft": "fort", "mt": "mount"}

var stateNames = map[string]string{
	"AL": "Alabama", "AK": "Alaska", "AZ": "Arizona", "AR": "Arkansas", "CA": "California",
	"CO": "Colorado", "CT": "Connecticut", "DE": "Delaware", "DC": "District of Columbia",
	"FL": "Florida", "GA": "Georgia", "HI": "Hawaii", "ID": "Idaho", "IL": "Illinois",
	"IN": "Indiana", "IA": "Iowa", "KS": "Kansas", "KY": "Kentucky", "LA": "Louisiana",
	"ME": "Maine", "MD": "Maryland", "MA": "Massachusetts", "MI": "Michigan", "MN": "Minnesota",
	"MS": "Mississippi", "MO": "Missouri", "MT": "Montana", "NE": "Nebraska", "NV": "Nevada",
	"NH": "New Hampshire", "NJ": "New Jersey", "NM": "New Mexico", "NY": "New York",
	"NC": "North Carolina", "ND": "North Dakota", "OH": "Ohio", "OK": "Oklahoma", "OR": "Oregon",
	"PA": "Pennsylvania", "RI": "Rhode Island", "SC": "South Carolina", "SD": "South Dakota",
	"TN": "Tennessee", "TX": "Texas", "UT": "Utah", "VT": "Vermont", "VA": "Virginia",
	"WA": "Washington", "WV": "West Virginia", "WI": "Wisconsin", "WY": "Wyoming",
}

// Preload loads the places and time zone boundaries in the background, so
// they're ready when first needed.
func Preload() {
	go func() {
		_, _ = Default()
		_, _ = loadFinder()
	}()
}

var loadFinder = sync.OnceValues(tzf.NewDefaultFinder)

// TimeZone returns the IANA time zone at the coordinates, such as
// "America/Chicago". The boundary data is loaded on first use.
func TimeZone(latitude, longitude float64) (*time.Location, error) {
	f, err := loadFinder()
	if err != nil {
		return nil, fmt.Errorf("place: time zones: %w", err)
	}
	name := f.GetTimezoneName(longitude, latitude)
	if name == "" {
		return nil, fmt.Errorf("place: no time zone at %.4f, %.4f", latitude, longitude)
	}
	return time.LoadLocation(name)
}
