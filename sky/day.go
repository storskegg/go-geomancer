package sky

import (
	"math"
	"slices"
	"time"

	"github.com/soniakeys/meeus/v3/base"
	"github.com/soniakeys/meeus/v3/coord"
	"github.com/soniakeys/meeus/v3/nutation"
	"github.com/soniakeys/meeus/v3/sidereal"
	"github.com/soniakeys/unit"

	"github.com/storskegg/go-geomancer/zodiac"
)

// Day is the planetary day and hour of a moment. The planetary day runs from
// sunrise to sunrise, so before sunrise it's still the previous day's. Its
// daylight and its night are each divided into twelve equal hours, ruled in
// Chaldean order from the day's ruler in the first.
type Day struct {
	// Ruler is the planet of the day's weekday: the Sun's on Sunday, the
	// Moon's on Monday, and so on.
	Ruler zodiac.Planet `json:"ruler"`
	// Sunrise began the planetary day; Sunset and NextSunrise end its
	// daylight and its night. Zero if the Sun doesn't rise or set.
	Sunrise     time.Time `json:"sunrise,omitzero"`
	Sunset      time.Time `json:"sunset,omitzero"`
	NextSunrise time.Time `json:"nextSunrise,omitzero"`
	// Hour is the planetary hour, or nil if the Sun doesn't rise or set
	// at the place that day, as can happen near the poles; Unavailable
	// then says which. The day then takes the calendar weekday's ruler.
	Hour        *Hour  `json:"hour,omitempty"`
	Unavailable string `json:"unavailable,omitempty"`
}

// Hour is a planetary hour.
type Hour struct {
	Ruler  zodiac.Planet `json:"ruler"`
	Number int           `json:"number"` // 1–12, of the day or of the night
	Night  bool          `json:"night,omitempty"`
	Start  time.Time     `json:"start"`
	End    time.Time     `json:"end"`
}

// chaldean is the planets in Chaldean order, slowest first, which rules
// the hours in turn.
var chaldean = []zodiac.Planet{zodiac.Saturn, zodiac.Jupiter, zodiac.Mars, zodiac.Sun, zodiac.Venus, zodiac.Mercury, zodiac.Moon}

// weekdayRulers gives the ruler of each weekday, Sunday first.
var weekdayRulers = [7]zodiac.Planet{zodiac.Sun, zodiac.Moon, zodiac.Mars, zodiac.Mercury, zodiac.Jupiter, zodiac.Venus, zodiac.Saturn}

// Reasons a planetary hour can't be worked out.
const (
	noSunset  = "the Sun doesn't set here on this date"
	noSunrise = "the Sun doesn't rise here on this date"
)

// planetaryDay works out the planetary day and hour at jd (UT), time t, at
// latitude and longitude lat, lon in degrees.
func planetaryDay(jd float64, t time.Time, lat, lon float64) Day {
	loc := t.Location()
	φ, east := lat*math.Pi/180, lon*math.Pi/180
	events := sunEvents(jd-1.5, jd+1.5, φ, east)

	var rise, set, next float64
	for _, e := range events {
		if e.rise && e.jd <= jd {
			rise = e.jd
		}
	}
	for _, e := range events {
		switch {
		case rise != 0 && set == 0 && !e.rise && e.jd > rise:
			set = e.jd
		case set != 0 && next == 0 && e.rise && e.jd > set:
			next = e.jd
		}
	}
	if rise == 0 || set == 0 || next == 0 {
		d := Day{Ruler: weekdayRulers[t.Weekday()], Unavailable: noSunrise}
		if sunAltitude(jd, φ, east) > horizon {
			d.Unavailable = noSunset
		}
		return d
	}

	d := Day{
		Sunrise:     fromJD(rise, loc),
		Sunset:      fromJD(set, loc),
		NextSunrise: fromJD(next, loc),
	}
	d.Ruler = weekdayRulers[d.Sunrise.Weekday()]
	start, length, n, night := rise, (set-rise)/12, 0, false
	if jd >= set {
		start, length, night = set, (next-set)/12, true
	}
	n = min(int((jd-start)/length), 11)
	first := slices.Index(chaldean, d.Ruler)
	offset := n
	if night {
		offset += 12
	}
	d.Hour = &Hour{
		Ruler:  chaldean[(first+offset)%7],
		Number: n + 1,
		Night:  night,
		Start:  fromJD(start+float64(n)*length, loc),
		End:    fromJD(start+float64(n+1)*length, loc),
	}
	return d
}

// horizon is the altitude of the Sun's centre at sunrise and sunset, in
// radians: the upper edge on the horizon, allowing for refraction (−0°50′).
const horizon = -50.0 / 60 * math.Pi / 180

type sunEvent struct {
	jd   float64
	rise bool
}

// sunEvents returns the sunrises and sunsets between jd0 and jd1, in order.
// It samples the Sun's altitude every ten minutes and narrows each crossing
// of the horizon to under a second.
func sunEvents(jd0, jd1, φ, east float64) []sunEvent {
	const step = 10.0 / 1440
	var out []sunEvent
	prevJD := jd0
	prev := sunAltitude(jd0, φ, east) - horizon
	for jd := jd0 + step; jd <= jd1; jd += step {
		cur := sunAltitude(jd, φ, east) - horizon
		if (prev < 0) != (cur < 0) {
			lo, hi := prevJD, jd
			for hi-lo > 0.5/86400 {
				mid := (lo + hi) / 2
				if (sunAltitude(mid, φ, east)-horizon < 0) == (prev < 0) {
					lo = mid
				} else {
					hi = mid
				}
			}
			out = append(out, sunEvent{jd: (lo + hi) / 2, rise: prev < 0})
		}
		prevJD, prev = jd, cur
	}
	return out
}

// sunAltitude returns the altitude of the Sun's centre, in radians, at jd
// (UT), seen from latitude φ and east longitude east, in radians.
func sunAltitude(jd, φ, east float64) float64 {
	jde := jd + deltaT(jd)
	Δψ, Δε := nutation.Nutation(jde)
	l0, b0, r0 := earth.position(jde)
	λ, β := toFK5(l0+math.Pi, -b0, (jde-base.J2000)/36525)
	λ += Δψ.Rad() - 20.4898*arcsec/r0
	ε := nutation.MeanObliquity(jde) + Δε
	α, δ := coord.EclToEq(unit.Angle(λ), unit.Angle(β), math.Sin(ε.Rad()), math.Cos(ε.Rad()))
	θ := sidereal.Apparent(jd).Rad() // Greenwich apparent sidereal time
	H := θ + east - α.Rad()
	return math.Asin(math.Sin(φ)*math.Sin(δ.Rad()) + math.Cos(φ)*math.Cos(δ.Rad())*math.Cos(H))
}
