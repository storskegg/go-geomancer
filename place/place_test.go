package place

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"strconv"
	"strings"
	"testing"
	"time"
)

func index(t *testing.T) *Index {
	t.Helper()
	ix, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	return ix
}

func labels(ms []Match) []string {
	out := make([]string, len(ms))
	for i, m := range ms {
		out[i] = m.Label
	}
	return out
}

func TestIndex(t *testing.T) {
	ix := index(t)
	if n := ix.Len(); n < 150000 {
		t.Errorf("%d places, want the wide coverage (over 150,000)", n)
	}
	states := map[string]bool{}
	for _, p := range ix.places {
		states[p.State] = true
		// The western Aleutians lie past the antimeridian, east of 172°E.
		aleutian := p.State == "AK" && p.Longitude > 172
		if p.Latitude < 18 || p.Latitude > 72 || (p.Longitude > -66 && !aleutian) || p.Longitude < -180 {
			t.Fatalf("%s is outside the US: %v, %v", p.Label(), p.Latitude, p.Longitude)
		}
	}
	if len(states) != 51 {
		t.Errorf("%d states, want the 50 and DC", len(states))
	}
	for i := 1; i < len(ix.places); i++ {
		if ix.places[i].Population > ix.places[i-1].Population {
			t.Fatalf("places aren't most populous first at %d", i)
		}
	}
}

func TestSearch(t *testing.T) {
	ix := index(t)
	tests := []struct {
		q     string
		first string // the first result
		all   string // a prefix every result has, if any
	}{
		{"springfield", "Springfield, MO", "Springfield"},
		{"Springfield, IL", "Springfield, IL", ""},
		{"springfield illinois", "Springfield, IL", ""},
		{"springfield sangamon", "Springfield, IL", ""},
		{"st louis", "St. Louis, MO", ""},
		{"saint louis", "St. Louis, MO", ""},
		{"canon city", "Cañon City, CO", ""},
		{"ft worth", "Fort Worth, TX", ""},
		{"ofallon il", "O'Fallon, IL", ""},
		{"new york", "New York City, NY", ""},
		{"york", "York, PA", ""},
		{"washington dc", "Washington, DC", ""},
		{"honolulu", "Honolulu, HI", ""},
	}
	for _, tt := range tests {
		got := labels(ix.Search(tt.q, 8))
		if len(got) == 0 || got[0] != tt.first {
			t.Errorf("Search(%q) = %v, want %s first", tt.q, got, tt.first)
		}
		for _, l := range got {
			if tt.all != "" && !strings.HasPrefix(l, tt.all) {
				t.Errorf("Search(%q) includes %s", tt.q, l)
			}
		}
	}

	// Duplicates come most populous first.
	got := ix.Search("springfield", 5)
	for i := 1; i < len(got); i++ {
		if got[i].Place.Population > got[i-1].Place.Population {
			t.Errorf("Springfields out of order: %v", labels(got))
		}
	}
	// A state abbreviation matches only the state.
	for _, m := range ix.Search("springfield pa", 20) {
		if m.Place.State != "PA" {
			t.Errorf("springfield pa found %s", m.Label)
		}
	}
	// The first word must begin the name.
	for _, m := range ix.Search("york", 50) {
		if !strings.HasPrefix(strings.ToLower(m.Place.Name), "york") {
			t.Errorf("york found %s", m.Label)
		}
	}
	if got := ix.Search("zzzzqq", 5); len(got) != 0 {
		t.Errorf("nonsense found %v", labels(got))
	}
	if got := ix.Search("  ", 5); got != nil {
		t.Errorf("empty query found %v", labels(got))
	}
	if got := ix.Search("springfield", 0); got != nil {
		t.Errorf("n = 0 found %v", labels(got))
	}
}

func TestLabels(t *testing.T) {
	ix := index(t)
	// Several Springfields in Pennsylvania are told apart by county.
	got := labels(ix.Search("springfield pa", 3))
	if len(got) < 2 || !strings.HasPrefix(got[0], "Springfield, PA (") || !strings.HasSuffix(got[0], "County)") {
		t.Errorf("Springfield, PA labels = %v", got)
	}
	// A unique name in its state has none.
	if m, ok := ix.Find("Springfield, IL"); !ok || m.Label != "Springfield, IL" {
		t.Errorf("Find(Springfield, IL) = %+v, %v", m, ok)
	}
	// A label with a county finds that place.
	if len(got) > 1 {
		m, ok := ix.Find(got[1])
		if !ok || m.Label != got[1] {
			t.Errorf("Find(%q) = %+v, %v", got[1], m, ok)
		}
	}
	for _, s := range []string{"springfield il", "SPRINGFIELD, Illinois", "Nowhere, ZZ"} {
		_, ok := ix.Find(s)
		if want := s == "springfield il"; ok != want {
			t.Errorf("Find(%q) ok = %v, want %v", s, ok, want)
		}
	}
}

func TestPlace(t *testing.T) {
	m, _ := index(t).Find("Springfield, IL")
	p := m.Place
	if p.Label() != "Springfield, IL" || p.County != "Sangamon County" || p.Population < 100000 {
		t.Errorf("Springfield, IL = %+v", p)
	}
	loc, err := TimeZone(p.Latitude, p.Longitude)
	if err != nil || loc.String() != "America/Chicago" {
		t.Fatalf("TimeZone = %v, %v", loc, err)
	}
	c := p.Circumstances(time.Date(2026, 10, 15, 15, 30, 0, 0, loc))
	if c.TimeZone != "America/Chicago" || c.Zone != "CDT" || c.Place.Label() != "Springfield, IL" || c.Place.Latitude != p.Latitude {
		t.Errorf("Circumstances = %+v", c)
	}
}

func TestTimeZones(t *testing.T) {
	for _, tt := range []struct {
		place string
		lat   float64
		lon   float64
		zone  string
	}{
		{"Phoenix", 33.44838, -112.07404, "America/Phoenix"},
		{"Indianapolis", 39.76838, -86.15804, "America/Indiana/Indianapolis"},
		{"El Paso", 31.75872, -106.48693, "America/Denver"},
		{"Honolulu", 21.30694, -157.85833, "Pacific/Honolulu"},
		{"Anchorage", 61.21806, -149.90028, "America/Anchorage"},
		{"Detroit", 42.33143, -83.04575, "America/Detroit"},
		{"Boise", 43.6135, -116.20345, "America/Boise"},
	} {
		loc, err := TimeZone(tt.lat, tt.lon)
		if err != nil || loc.String() != tt.zone {
			t.Errorf("%s: %v, %v; want %s", tt.place, loc, err, tt.zone)
		}
	}
}

// The time zone found from each place's coordinates agrees with the one
// GeoNames records for it, for a spread of places.
func TestTimeZonesAgreeWithGeoNames(t *testing.T) {
	r, err := gzip.NewReader(bytes.NewReader(placeFile))
	if err != nil {
		t.Fatal(err)
	}
	sc := bufio.NewScanner(r)
	checked, differ := 0, 0
	for i := 0; sc.Scan(); i++ {
		if i%56 != 0 { // about 3,000 places, populous and small
			continue
		}
		f := strings.Split(sc.Text(), "\t")
		lat, _ := strconv.ParseFloat(f[4], 64)
		lon, _ := strconv.ParseFloat(f[5], 64)
		loc, err := TimeZone(lat, lon)
		if err != nil {
			t.Fatalf("%s, %s: %v", f[0], f[2], err)
		}
		checked++
		if loc.String() != f[7] && !sameRules(loc, f[7]) {
			differ++
			if differ <= 5 {
				t.Logf("%s, %s: found %s, GeoNames has %s", f[0], f[2], loc, f[7])
			}
		}
	}
	t.Logf("%d of %d places differ", differ, checked)
	if float64(differ) > 0.01*float64(checked) {
		t.Errorf("%d of %d places disagree with GeoNames, want under 1%%", differ, checked)
	}
}

// sameRules reports whether loc and the named zone keep the same clocks, as
// aliases and neighbouring zones often do (America/Indiana/Indianapolis and
// America/Indianapolis).
func sameRules(loc *time.Location, name string) bool {
	other, err := time.LoadLocation(name)
	if err != nil {
		return false
	}
	for m := time.January; m <= time.December; m++ {
		t := time.Date(2026, m, 15, 12, 0, 0, 0, time.UTC)
		a, oa := t.In(loc).Zone()
		b, ob := t.In(other).Zone()
		if a != b || oa != ob {
			return false
		}
	}
	return true
}

func TestParseInvalid(t *testing.T) {
	gz := func(s string) []byte {
		var b bytes.Buffer
		w := gzip.NewWriter(&b)
		_, _ = w.Write([]byte(s))
		_ = w.Close()
		return b.Bytes()
	}
	for name, data := range map[string][]byte{
		"not gzip":       []byte("plain"),
		"too few fields": gz("Springfield\tSpringfield\tIL\n"),
		"bad latitude":   gz("A\tA\tIL\tC\tnorth\t-89\t1\tAmerica/Chicago\n"),
	} {
		if _, err := parse(data); err == nil {
			t.Errorf("%s: parsed", name)
		}
	}
}
