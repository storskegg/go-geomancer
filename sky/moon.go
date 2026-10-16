package sky

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/soniakeys/meeus/v3/base"
	"github.com/soniakeys/meeus/v3/moonphase"

	"github.com/storskegg/go-geomancer/zodiac"
)

// Phase is one of the Moon's eight phases, each 45° of its elongation from
// the Sun, beginning at the conjunction, as astrologers divide them.
type Phase uint8

const (
	NewMoon Phase = iota
	Crescent
	FirstQuarter
	Gibbous
	FullMoon
	Disseminating
	LastQuarter
	Balsamic
)

var phaseLabels = []string{"New", "Crescent", "First Quarter", "Gibbous", "Full", "Disseminating", "Last Quarter", "Balsamic"}

func (p Phase) String() string                { return label(phaseLabels, "Phase", p) }
func (p Phase) MarshalText() ([]byte, error)  { return marshal(phaseLabels, "phase", p) }
func (p *Phase) UnmarshalText(b []byte) error { return unmarshal(phaseLabels, "phase", b, p) }

// Waxing reports whether the Moon grows in this phase: from New to Full.
func (p Phase) Waxing() bool { return p < FullMoon }

// Moon is the Moon's phase and the lunations around it.
type Moon struct {
	Phase Phase `json:"phase"`
	// Elongation is the Moon's longitude less the Sun's, in degrees
	// [0, 360).
	Elongation float64 `json:"elongation"`
	// Illuminated is the fraction of the disc lit, 0 to 1.
	Illuminated float64 `json:"illuminated"`
	// The New and Full Moons before and after the moment, in its zone.
	PreviousNew  time.Time `json:"previousNew"`
	PreviousFull time.Time `json:"previousFull"`
	NextNew      time.Time `json:"nextNew"`
	NextFull     time.Time `json:"nextFull"`
	// LeavesSign is when the Moon next enters a sign.
	LeavesSign time.Time `json:"leavesSign"`
	// NextAspect is the first Ptolemaic aspect the Moon perfects with a
	// planet before it leaves its sign, or nil if it perfects none: it's
	// then void of course.
	NextAspect   *MoonAspect `json:"nextAspect,omitempty"`
	VoidOfCourse bool        `json:"voidOfCourse,omitempty"`
}

// MoonAspect is an aspect the Moon will perfect.
type MoonAspect struct {
	Planet zodiac.Planet `json:"planet"`
	Kind   Kind          `json:"kind"`
	Exact  time.Time     `json:"exact"`
}

const synodicMonth = 29.530588861 // days

// moon works out the phase at jde from the longitudes, the Sun's distance in
// au and the Moon's in km, and the lunations either side, in loc.
func moon(jde float64, lon [9]float64, sunAU, moonKm float64, loc *time.Location) Moon {
	e := norm(lon[zodiac.Moon] - lon[zodiac.Sun])
	// The phase angle, Sun–Moon–Earth, from the elongation and distances
	// (Meeus 48.3), ignoring the Moon's latitude.
	sunKm := sunAU * base.AU
	ψ := angle(lon[zodiac.Moon], lon[zodiac.Sun])
	i := math.Atan2(sunKm*math.Sin(ψ), moonKm-sunKm*math.Cos(ψ))
	m := Moon{
		Phase:       Phase(int(e/(math.Pi/4)) % 8),
		Elongation:  e * 180 / math.Pi,
		Illuminated: (1 + math.Cos(i)) / 2,
	}
	m.PreviousNew, m.NextNew = around(jde, moonphase.New, loc)
	m.PreviousFull, m.NextFull = around(jde, moonphase.Full, loc)
	leaves, next := moonCourse(jde, lon)
	m.LeavesSign = fromJD(leaves-deltaT(leaves), loc)
	if next != nil {
		m.NextAspect = &MoonAspect{Planet: next.planet, Kind: next.kind, Exact: fromJD(next.jde-deltaT(next.jde), loc)}
	} else {
		m.VoidOfCourse = true
	}
	return m
}

type perfection struct {
	planet zodiac.Planet
	kind   Kind
	jde    float64
}

// moonCourse follows the Moon from jde, when the bodies are at lon, until it
// leaves its sign, and returns when it does and the first Ptolemaic aspect
// it perfects with a planet on the way, if any. It samples hourly, which the
// Moon's half-degree an hour makes safe, and narrows each event to under a
// minute. The Moon is faster than every planet, so its distance past each
// exact aspect only grows: each perfection is a crossing of zero from below.
func moonCourse(jde float64, lon [9]float64) (leaves float64, first *perfection) {
	const step, maxDays = 1.0 / 24, 3.0 // the Moon crosses a sign in under 2.7 days
	cusp := math.Floor(lon[zodiac.Moon]/(math.Pi/6)+1) * math.Pi / 6
	pastCusp := func(l [9]float64) float64 { return signedDiff(l[zodiac.Moon], cusp) }

	type target struct {
		planet zodiac.Planet
		kind   Kind
		offset float64 // the Moon's longitude less the planet's when exact
	}
	var targets []target
	for _, p := range planets {
		if p == zodiac.Moon {
			continue
		}
		for k, info := range kindInfo {
			a := info.angle * math.Pi / 180
			targets = append(targets, target{p, Kind(k), a})
			if a != 0 && a != math.Pi {
				targets = append(targets, target{p, Kind(k), -a})
			}
		}
	}
	pastExact := func(tg target) func([9]float64) float64 {
		return func(l [9]float64) float64 { return signedDiff(l[zodiac.Moon]-l[tg.planet], tg.offset) }
	}

	prevJD, prevCusp := jde, pastCusp(lon)
	prev := make([]float64, len(targets))
	for i, tg := range targets {
		prev[i] = pastExact(tg)(lon)
	}
	for t := jde + step; t <= jde+maxDays; t += step {
		cur, _, _ := longitudes(t)
		left := 0.0
		if c := pastCusp(cur); prevCusp < 0 && c >= 0 {
			left = narrow(prevJD, t, pastCusp)
		}
		for i, tg := range targets {
			v := pastExact(tg)(cur)
			// A crossing from below, not the jump at ±π opposite it.
			if prev[i] < 0 && v >= 0 && v-prev[i] < math.Pi/2 {
				at := narrow(prevJD, t, pastExact(tg))
				if (left == 0 || at < left) && (first == nil || at < first.jde) {
					first = &perfection{tg.planet, tg.kind, at}
				}
			}
			prev[i] = v
		}
		if left != 0 {
			return left, first
		}
		prevJD, prevCusp = t, pastCusp(cur)
	}
	return jde + maxDays, first
}

// narrow finds, to under a minute, when f of the longitudes crosses zero
// between lo, where it's negative, and hi.
func narrow(lo, hi float64, f func([9]float64) float64) float64 {
	for hi-lo > 30.0/86400 {
		mid := (lo + hi) / 2
		if l, _, _ := longitudes(mid); f(l) < 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// around returns the lunations before and after jde, given a function that
// finds the one nearest a decimal year.
func around(jde float64, nearest func(year float64) float64, loc *time.Location) (before, after time.Time) {
	prev, next := math.Inf(-1), math.Inf(1)
	for _, d := range []float64{-synodicMonth, 0, synodicMonth} {
		l := nearest(2000 + (jde+d-base.J2000)/365.25)
		if l <= jde && l > prev {
			prev = l
		}
		if l > jde && l < next {
			next = l
		}
	}
	toTime := func(jde float64) time.Time { return fromJD(jde-deltaT(jde), loc) }
	return toTime(prev), toTime(next)
}

// label and marshal support the package's enums.

func label[T ~uint8](labels []string, kind string, v T) string {
	if int(v) < len(labels) {
		return labels[v]
	}
	return fmt.Sprintf("%s(%d)", kind, v)
}

func marshal[T ~uint8](labels []string, kind string, v T) ([]byte, error) {
	if int(v) >= len(labels) {
		return nil, fmt.Errorf("sky: unknown %s %d", kind, v)
	}
	return []byte(labels[v]), nil
}

func unmarshal[T ~uint8](labels []string, kind string, b []byte, v *T) error {
	i := slices.Index(labels, string(b))
	if i < 0 {
		return fmt.Errorf("sky: unknown %s %q", kind, b)
	}
	*v = T(i)
	return nil
}
