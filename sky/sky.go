// Package sky works out the astrology of a moment at a place: where the Sun,
// Moon, planets and lunar nodes are in the tropical zodiac, their dignities,
// their closeness to the Sun and their aspects to one another, the Moon's
// phase and lunations, and the planetary day and hour.
//
// Positions are apparent geocentric ecliptic longitudes. The Sun and planets
// come from the VSOP87D theory (Bretagnon and Francou, 1988), truncated to
// well under an arcsecond between 1800 and 2200; the Moon and its mean node
// from Meeus's Astronomical Algorithms, via github.com/soniakeys/meeus.
package sky

import (
	"fmt"
	"math"
	"time"

	"github.com/storskegg/go-geomancer/zodiac"
)

// Sky is the astrology of a moment.
type Sky struct {
	// Time is the moment, in the zone it was given in.
	Time time.Time `json:"time"`
	// Bodies are the planets and nodes, in zodiac.Planets order.
	Bodies []Body `json:"bodies"`
	// Aspects between the planets, and conjunctions with the nodes, closest
	// first.
	Aspects []Aspect `json:"aspects"`
	Moon    Moon     `json:"moon"`
	Day     Day      `json:"day"`
}

// Body is where a planet or node is.
type Body struct {
	Planet zodiac.Planet `json:"planet"`
	// Longitude is the apparent geocentric ecliptic longitude, in degrees
	// [0, 360) from 0° Aries.
	Longitude float64     `json:"longitude"`
	Sign      zodiac.Sign `json:"sign"`
	// Speed is the daily motion in longitude, in degrees; negative is
	// retrograde.
	Speed float64 `json:"speed"`
	// Retrograde is set for a planet moving backwards. The mean nodes
	// always do, so it isn't set for them.
	Retrograde bool `json:"retrograde,omitempty"`
	// Dignities are the planet's essential dignities and debilities by sign:
	// Peregrine if none. None for the nodes.
	Dignities []Dignity `json:"dignities,omitempty"`
	// Solar is the planet's closeness to the Sun, if close: never for the
	// Sun or the nodes.
	Solar Solar `json:"solar,omitempty"`
}

// Position writes the body's place in its sign, to the minute, truncated as
// astrologers do: "14°22′ Scorpio".
func (b Body) Position() string {
	d, m := degreesMinutes(b.Longitude)
	return fmt.Sprintf("%d°%02d′ %s", d, m, b.Sign)
}

// degreesMinutes returns the whole degrees and minutes of a longitude within
// its sign.
func degreesMinutes(lon float64) (int, int) {
	in := math.Mod(lon, 30)
	d := int(in)
	return d, int((in - float64(d)) * 60)
}

// Body returns the body for p.
func (s *Sky) Body(p zodiac.Planet) Body { return s.Bodies[p] }

// speedStep is the interval, in days, either side of the moment over which
// speeds are measured.
const speedStep = 1.0 / 24

// At returns the sky at time t, seen from latitude lat and longitude lon in
// degrees (north and east positive). Times in the result are in t's zone.
func At(t time.Time, lat, lon float64) (*Sky, error) {
	if math.Abs(lat) > 90 || math.Abs(lon) > 180 {
		return nil, fmt.Errorf("sky: invalid coordinates %v, %v", lat, lon)
	}
	jd := toJD(t)
	jde := jd + deltaT(jd)
	now, sunDist, moonDist := longitudes(jde)
	before, _, _ := longitudes(jde - speedStep)
	after, _, _ := longitudes(jde + speedStep)

	s := &Sky{Time: t}
	for _, p := range zodiac.Planets() {
		b := Body{
			Planet:    p,
			Longitude: now[p] * 180 / math.Pi,
			Speed:     signedDiff(after[p], before[p]) * 180 / math.Pi / (2 * speedStep),
		}
		b.Sign = zodiac.Sign(int(b.Longitude/30) % 12)
		if p != zodiac.NorthNode && p != zodiac.SouthNode {
			b.Retrograde = b.Speed < 0
			b.Dignities = dignities(p, b.Sign)
			if p != zodiac.Sun {
				b.Solar = solar(angle(now[p], now[zodiac.Sun]))
			}
		}
		s.Bodies = append(s.Bodies, b)
	}
	s.Aspects = aspects(now, after)
	s.Moon = moon(jde, now, sunDist, moonDist, t.Location())
	s.Day = planetaryDay(jd, t, lat, lon)
	return s, nil
}

// signedDiff returns a − b, in radians (−π, π].
func signedDiff(a, b float64) float64 {
	d := math.Mod(a-b, 2*math.Pi)
	switch {
	case d > math.Pi:
		d -= 2 * math.Pi
	case d <= -math.Pi:
		d += 2 * math.Pi
	}
	return d
}

// angle returns the angular distance between two longitudes, in radians
// [0, π].
func angle(a, b float64) float64 { return math.Abs(signedDiff(a, b)) }
