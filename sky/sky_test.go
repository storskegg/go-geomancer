package sky

import (
	"encoding/json"
	"math"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/storskegg/go-geomancer/zodiac"
)

func chicago(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

const madisonLat, madisonLon = 43.0731, -89.4012

func madison(t *testing.T, tm time.Time) *Sky {
	t.Helper()
	s, err := At(tm, madisonLat, madisonLon)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// The truncated series against the check values published with VSOP87D
// (CDS VI/81, vsop87.chk), at J2000 and J1900.
func TestVSOP87(t *testing.T) {
	tests := []struct {
		name    string
		s       *vsop87
		jde     float64
		l, b, r float64
	}{
		{"mercury", &mercury, 2451545.0, 4.4293481036, -.0527573409, .4664714751},
		{"mercury", &mercury, 2415020.0, 3.4851161911, .0565906173, .4183426275},
		{"venus", &venus, 2451545.0, 3.1870221833, .0569782849, .7202129253},
		{"venus", &venus, 2415020.0, 5.9749622238, -.0591260014, .7274719359},
		{"earth", &earth, 2451545.0, 1.7519238681, -.0000039656, .9833276819},
		{"earth", &earth, 2415020.0, 1.7391225563, -.0000005679, .9832689778},
		{"mars", &mars, 2451545.0, 6.2735389983, -.0247779824, 1.3912076925},
		{"mars", &mars, 2415020.0, 4.9942005211, -.0271965869, 1.4218777705},
		{"jupiter", &jupiter, 2451545.0, .6334614186, -.0205001039, 4.9653813154},
		{"jupiter", &jupiter, 2415020.0, 4.0927527024, .0161446618, 5.3850276671},
		{"saturn", &saturn, 2451545.0, .7980038761, -.0401984149, 9.1838483715},
		{"saturn", &saturn, 2415020.0, 4.6512836347, .0192701409, 10.0668531997},
	}
	const maxAngle, maxDist = 1 * arcsec, 1e-5 // radians, au
	for _, tt := range tests {
		l, b, r := tt.s.position(tt.jde)
		if d := angle(l, tt.l); d > maxAngle || math.Abs(b-tt.b) > maxAngle || math.Abs(r-tt.r) > maxDist {
			t.Errorf("%s at JD %v: l %v b %v r %v; want %v %v %v", tt.name, tt.jde, norm(l), b, r, tt.l, tt.b, tt.r)
		}
	}
}

// Apparent geocentric longitudes against JPL Horizons (observer table,
// quantity 31: ecliptic of date), for 2026-10-15 18:14 UT.
func TestHorizons(t *testing.T) {
	want := map[zodiac.Planet]float64{
		zodiac.Sun:     202.4099271,
		zodiac.Moon:    260.7275509,
		zodiac.Mercury: 227.1042314,
		zodiac.Venus:   215.4692232,
		zodiac.Mars:    130.0552437,
		zodiac.Jupiter: 142.0661601,
		zodiac.Saturn:  10.4275590,
	}
	s := madison(t, time.Date(2026, 10, 15, 18, 14, 0, 0, time.UTC))
	for p, lon := range want {
		if d := angle(s.Body(p).Longitude*math.Pi/180, lon*math.Pi/180) / arcsec; d > 10 {
			t.Errorf("%v: %.5f°, Horizons %.5f° (%.1f″ off)", p, s.Body(p).Longitude, lon, d)
		}
	}
}

func TestBodies(t *testing.T) {
	s := madison(t, time.Date(2026, 10, 15, 13, 14, 0, 0, chicago(t)))
	tests := []struct {
		p          zodiac.Planet
		position   string
		retrograde bool
		dignities  []Dignity
		solar      Solar
	}{
		{zodiac.Sun, "22°24′ Libra", false, []Dignity{Fall}, NotNearSun},
		{zodiac.Moon, "20°43′ Sagittarius", false, []Dignity{Peregrine}, NotNearSun},
		{zodiac.Venus, "5°28′ Scorpio", true, []Dignity{Detriment}, UnderTheBeams},
		{zodiac.Saturn, "10°25′ Aries", true, []Dignity{Fall}, NotNearSun},
		{zodiac.NorthNode, "26°55′ Aquarius", false, nil, NotNearSun},
		{zodiac.SouthNode, "26°55′ Leo", false, nil, NotNearSun},
	}
	for _, tt := range tests {
		b := s.Body(tt.p)
		if b.Planet != tt.p || b.Position() != tt.position || b.Retrograde != tt.retrograde || !slices.Equal(b.Dignities, tt.dignities) || b.Solar != tt.solar {
			t.Errorf("%v: %s, retrograde %v, %v, %q; want %s, %v, %v, %q",
				tt.p, b.Position(), b.Retrograde, b.Dignities, b.Solar, tt.position, tt.retrograde, tt.dignities, tt.solar)
		}
	}
}

func TestPosition(t *testing.T) {
	for lon, want := range map[float64]string{
		0:        "0°00′ Aries",
		224.9999: "14°59′ Scorpio", // truncated, not rounded
		359.99:   "29°59′ Pisces",
		90.5:     "0°30′ Cancer",
	} {
		b := Body{Longitude: lon, Sign: zodiac.Sign(int(lon / 30))}
		if got := b.Position(); got != want {
			t.Errorf("Position(%v) = %q, want %q", lon, got, want)
		}
	}
}

func TestDignities(t *testing.T) {
	tests := []struct {
		p    zodiac.Planet
		s    zodiac.Sign
		want []Dignity
	}{
		{zodiac.Mercury, zodiac.Virgo, []Dignity{Domicile, Exaltation}},
		{zodiac.Mercury, zodiac.Pisces, []Dignity{Detriment, Fall}},
		{zodiac.Moon, zodiac.Cancer, []Dignity{Domicile}},
		{zodiac.Moon, zodiac.Scorpio, []Dignity{Fall}},
		{zodiac.Mars, zodiac.Capricorn, []Dignity{Exaltation}},
		{zodiac.Venus, zodiac.Aries, []Dignity{Detriment}},
		{zodiac.Sun, zodiac.Aquarius, []Dignity{Detriment}},
		{zodiac.Jupiter, zodiac.Leo, []Dignity{Peregrine}},
		{zodiac.Saturn, zodiac.Aquarius, []Dignity{Domicile}},
	}
	for _, tt := range tests {
		if got := dignities(tt.p, tt.s); !slices.Equal(got, tt.want) {
			t.Errorf("%v in %v: %v, want %v", tt.p, tt.s, got, tt.want)
		}
	}
}

func TestSolar(t *testing.T) {
	for deg, want := range map[float64]Solar{
		0.2: Cazimi, 0.29: Combust, 8.5: Combust, 8.6: UnderTheBeams, 17: UnderTheBeams, 17.1: NotNearSun,
	} {
		if got := solar(deg * math.Pi / 180); got != want {
			t.Errorf("solar(%v°) = %q, want %q", deg, got, want)
		}
	}
}

func TestAspects(t *testing.T) {
	rad := func(deg float64) float64 { return deg * math.Pi / 180 }
	var now, later [9]float64
	// Sun at 0°, Moon at 95° and moving away from the square; Mars at 245°,
	// trine the Sun within 5° and closing; Saturn opposite Mars by 11°,
	// beyond their 8°15′; Venus on the North Node at 30°.
	set := func(p zodiac.Planet, n, l float64) { now[p], later[p] = rad(n), rad(l) }
	set(zodiac.Sun, 0, 0)
	set(zodiac.Moon, 95, 96)
	set(zodiac.Mercury, 300, 300)
	set(zodiac.Venus, 30, 30)
	set(zodiac.Mars, 245, 244)
	set(zodiac.Jupiter, 200, 200)
	set(zodiac.Saturn, 76, 76)
	set(zodiac.NorthNode, 31, 31)
	set(zodiac.SouthNode, 211, 211)

	got := map[[2]zodiac.Planet]Aspect{}
	for _, a := range aspects(now, later) {
		got[[2]zodiac.Planet{a.A, a.B}] = a
	}
	type want struct {
		kind     Kind
		orb      float64
		applying bool
	}
	for pair, w := range map[[2]zodiac.Planet]want{
		{zodiac.Sun, zodiac.Moon}:        {Square, 5, false},
		{zodiac.Sun, zodiac.Mars}:        {Trine, 5, true},
		{zodiac.Venus, zodiac.NorthNode}: {Conjunction, 1, false},
	} {
		a, ok := got[pair]
		if !ok || a.Kind != w.kind || math.Abs(a.Orb-w.orb) > 1e-9 || a.Applying != w.applying {
			t.Errorf("%v: %+v (found %v), want %+v", pair, a, ok, w)
		}
	}
	if a, ok := got[[2]zodiac.Planet{zodiac.Mars, zodiac.Saturn}]; ok {
		t.Errorf("Mars and Saturn 11° from opposition: %+v", a)
	}
	// The nodes make only conjunctions: Saturn is sextile the North Node
	// within 1°, which doesn't count.
	if a, ok := got[[2]zodiac.Planet{zodiac.Saturn, zodiac.NorthNode}]; ok {
		t.Errorf("Saturn and the North Node: %+v", a)
	}
}

func TestPlanetaryDay(t *testing.T) {
	loc := chicago(t)
	at := func(h, m int) *Sky { return madison(t, time.Date(2026, 10, 15, h, m, 0, 0, loc)) }

	// Thursday, after sunrise: the day of Jupiter, whose hour is first.
	s := at(13, 14)
	d := s.Day
	if d.Ruler != zodiac.Jupiter || d.Hour == nil || d.Hour.Ruler != zodiac.Saturn || d.Hour.Number != 7 || d.Hour.Night {
		t.Fatalf("13:14: %+v, hour %+v", d, d.Hour)
	}
	// Sunrise and sunset against JPL Horizons (true visual horizon, to the
	// minute): 12:12 and 23:15 UT.
	for name, c := range map[string][2]time.Time{
		"sunrise": {d.Sunrise, time.Date(2026, 10, 15, 12, 12, 0, 0, time.UTC)},
		"sunset":  {d.Sunset, time.Date(2026, 10, 15, 23, 15, 0, 0, time.UTC)},
	} {
		if diff := c[0].Sub(c[1]); diff < -90*time.Second || diff > 90*time.Second {
			t.Errorf("%s %v, Horizons %v", name, c[0], c[1])
		}
	}
	if first := at(7, 12).Day.Hour; first.Ruler != zodiac.Jupiter || first.Number != 1 || first.Night {
		t.Errorf("first hour %+v", first)
	}

	// Before sunrise it's still Wednesday, Mercury's day, in its night.
	// Counting on from Mercury in Chaldean order, the 11th night hour (the
	// 23rd) is the Moon's and the 12th Saturn's, which leads into Thursday's
	// first, Jupiter's.
	if d := at(5, 30).Day; d.Ruler != zodiac.Mercury || d.Hour == nil || !d.Hour.Night || d.Hour.Ruler != zodiac.Moon || d.Hour.Number != 11 {
		t.Errorf("05:30: %+v, hour %+v", d, d.Hour)
	}
	if h := at(6, 30).Day.Hour; !h.Night || h.Number != 12 || h.Ruler != zodiac.Saturn {
		t.Errorf("06:30: %+v", h)
	}
	// After sunset: Thursday's 12th day hour is Mercury's, so the 1st night
	// hour is the Moon's.
	if h := at(18, 30).Day.Hour; !h.Night || h.Number != 1 || h.Ruler != zodiac.Moon {
		t.Errorf("18:30: %+v", h)
	}

	// Hours are equal: each a twelfth of the daylight or the night, across
	// the end of daylight saving on 1 November.
	night := madison(t, time.Date(2026, 10, 31, 23, 0, 0, 0, loc)).Day
	length := night.NextSunrise.Sub(night.Sunset) / 12
	if got := night.Hour.End.Sub(night.Hour.Start); (got - length).Abs() > time.Second {
		t.Errorf("hour across the clock change lasts %v, want %v", got, length)
	}
}

// At Utqiaġvik the Sun doesn't set at midsummer or rise at midwinter: the
// day takes the weekday's ruler and the hour is unavailable.
func TestPolar(t *testing.T) {
	ak, err := time.LoadLocation("America/Anchorage")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		date   time.Time
		ruler  zodiac.Planet
		reason string
	}{
		{time.Date(2026, 6, 21, 12, 0, 0, 0, ak), zodiac.Sun, noSunset},    // a Sunday
		{time.Date(2026, 12, 21, 12, 0, 0, 0, ak), zodiac.Moon, noSunrise}, // a Monday
	}
	for _, tt := range tests {
		s, err := At(tt.date, 71.2906, -156.7886)
		if err != nil {
			t.Fatal(err)
		}
		if d := s.Day; d.Ruler != tt.ruler || d.Hour != nil || d.Unavailable != tt.reason {
			t.Errorf("%v: %+v", tt.date, d)
		}
	}
}

// The lunations are the conjunction and opposition of the Sun and Moon, on
// either side of the moment.
func TestLunations(t *testing.T) {
	tm := time.Date(2026, 10, 15, 13, 14, 0, 0, chicago(t))
	m := madison(t, tm).Moon
	if m.Phase != Crescent || math.Abs(m.Illuminated-0.238) > 0.005 || !m.Phase.Waxing() {
		t.Errorf("phase %v, %.3f lit", m.Phase, m.Illuminated)
	}
	if m.PreviousNew.After(tm) || m.PreviousFull.After(tm) || m.NextNew.Before(tm) || m.NextFull.Before(tm) {
		t.Fatalf("lunations not around the moment: %+v", m)
	}
	if month := m.NextNew.Sub(m.PreviousNew).Hours() / 24; month < 29.2 || month > 29.9 {
		t.Errorf("new Moons %v apart", month)
	}
	for _, c := range []struct {
		at   time.Time
		want float64
	}{{m.PreviousNew, 0}, {m.NextNew, 0}, {m.PreviousFull, 180}, {m.NextFull, 180}} {
		e := madison(t, c.at).Moon.Elongation
		if d := angle(e*math.Pi/180, c.want*math.Pi/180) * 180 / math.Pi; d > 0.02 {
			t.Errorf("elongation at %v = %.3f°, want %v°", c.at, e, c.want)
		}
	}
	if m.PreviousNew.Location() != tm.Location() {
		t.Errorf("lunations in %v, want %v", m.PreviousNew.Location(), tm.Location())
	}
}

// The Moon's course: its next aspect, exact when its distance from the
// planet is the aspect's angle, and void of course once it has made its
// last aspect in a sign.
func TestMoonCourse(t *testing.T) {
	s := madison(t, time.Date(2026, 10, 15, 18, 14, 0, 0, time.UTC))
	m := s.Moon
	if m.VoidOfCourse || m.NextAspect == nil || m.NextAspect.Planet != zodiac.Jupiter || m.NextAspect.Kind != Trine {
		t.Fatalf("next aspect %+v, void %v; want a trine to Jupiter", m.NextAspect, m.VoidOfCourse)
	}
	exact := madison(t, m.NextAspect.Exact)
	if d := angle(exact.Body(zodiac.Moon).Longitude*math.Pi/180, exact.Body(zodiac.Jupiter).Longitude*math.Pi/180) * 180 / math.Pi; math.Abs(d-120) > 0.02 {
		t.Errorf("at %v the Moon is %.3f° from Jupiter, want 120°", m.NextAspect.Exact, d)
	}
	// It enters Capricorn at the cusp.
	leaves := m.LeavesSign
	before, after := madison(t, leaves.Add(-time.Minute)).Body(zodiac.Moon), madison(t, leaves.Add(time.Minute)).Body(zodiac.Moon)
	if before.Sign != zodiac.Sagittarius || after.Sign != zodiac.Capricorn {
		t.Errorf("leaves %v: %v a minute before, %v after", leaves, before.Position(), after.Position())
	}

	// After its sextile to the Sun, it makes no more aspects in
	// Sagittarius.
	v := madison(t, time.Date(2026, 10, 16, 6, 0, 0, 0, time.UTC)).Moon
	if !v.VoidOfCourse || v.NextAspect != nil || v.LeavesSign.Sub(leaves).Abs() > time.Minute {
		t.Errorf("void %v, next %+v, leaves %v; want void until %v", v.VoidOfCourse, v.NextAspect, v.LeavesSign, leaves)
	}
}

func TestJSON(t *testing.T) {
	b, err := json.Marshal(madison(t, time.Date(2026, 10, 15, 13, 14, 0, 0, chicago(t))))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"planet":"Venus"`, `"retrograde":true`, `"dignities":["Detriment"]`, `"solar":"Under the beams"`,
		`"bodies":["Mars","Saturn"],"kind":"Trine"`, `"phase":"Crescent"`, `"ruler":"Jupiter"`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("JSON lacks %s:\n%s", want, b)
		}
	}
}

// A sky survives a round trip through JSON.
func TestJSONRoundTrip(t *testing.T) {
	s := madison(t, time.Date(2026, 10, 15, 13, 14, 0, 0, chicago(t)))
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var got Sky
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	again, err := json.Marshal(&got)
	if err != nil {
		t.Fatal(err)
	}
	if string(again) != string(b) {
		t.Errorf("round trip changed the JSON:\n%s\n%s", b, again)
	}
}

func TestInvalidCoordinates(t *testing.T) {
	if _, err := At(time.Now(), 91, 0); err == nil {
		t.Error("latitude 91 accepted")
	}
}
