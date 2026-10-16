package sky

import (
	"math"
	"time"

	"github.com/soniakeys/meeus/v3/base"
	"github.com/soniakeys/meeus/v3/deltat"
	"github.com/soniakeys/meeus/v3/julian"
	"github.com/soniakeys/meeus/v3/moonposition"
	"github.com/soniakeys/meeus/v3/nutation"

	"github.com/storskegg/go-geomancer/zodiac"
)

// Times are Julian days: jd in Universal Time, for the Earth's rotation, and
// jde in Terrestrial Time, for the ephemerides. They differ by ΔT.

// toJD returns the Julian day (UT) of t.
func toJD(t time.Time) float64 { return julian.TimeToJD(t.UTC()) } // TimeToJD ignores the zone

// fromJD returns the time of Julian day jd (UT) in loc, to the second.
func fromJD(jd float64, loc *time.Location) time.Time {
	return julian.JDToTime(jd).Round(time.Second).In(loc)
}

// deltaT returns TT − UT in days at Julian day jd. From 2005 it uses
// Espenak and Meeus's polynomials (NASA, Five Millennium Canon of Solar
// Eclipses, 2006), and before then the table of observed values, from 1620,
// or Meeus's polynomials.
func deltaT(jd float64) float64 {
	y := 2000 + (jd-base.J2000)/365.25
	var s float64
	switch {
	case y >= 2150:
		u := (y - 1820) / 100
		s = -20 + 32*u*u
	case y >= 2050:
		u := (y - 1820) / 100
		s = -20 + 32*u*u - 0.5628*(2150-y)
	case y >= 2005:
		t := y - 2000
		s = 62.92 + 0.32217*t + 0.005589*t*t
	case y >= 1620:
		s = deltat.Interp10A(jd).Sec()
	case y >= 948:
		s = deltat.Poly948to1600(y).Sec()
	default:
		s = deltat.PolyBefore948(y).Sec()
	}
	return s / 86400
}

// vsop87 is a planet's VSOP87D series: for its heliocentric longitude,
// latitude and distance, the terms multiplying each power of T, in Julian
// millennia from J2000 (TDB, taken as TT).
type vsop87 [3][][]vsopTerm

// vsopTerm is A cos(B + C T).
type vsopTerm struct{ a, b, c float64 }

// position returns the heliocentric ecliptic longitude and latitude (radians)
// and distance (au) at jde, referred to the ecliptic and equinox of date.
func (s *vsop87) position(jde float64) (l, b, r float64) {
	t := (jde - base.J2000) / 365250
	var out [3]float64
	for v := range s {
		tk := 1.0
		for _, terms := range s[v] {
			x := 0.0
			for _, tm := range terms {
				x += tm.a * math.Cos(tm.b+tm.c*t)
			}
			out[v] += x * tk
			tk *= t
		}
	}
	return out[0], out[1], out[2]
}

var series = map[zodiac.Planet]*vsop87{
	zodiac.Mercury: &mercury, zodiac.Venus: &venus, zodiac.Mars: &mars, zodiac.Jupiter: &jupiter, zodiac.Saturn: &saturn,
}

const arcsec = math.Pi / 180 / 3600

// longitudes returns the apparent geocentric ecliptic longitude of each
// planet at jde, in radians [0, 2π), and the distances of the Sun, in au,
// and the Moon, in km, which the Moon's phase needs. The nodes are the mean
// nodes.
func longitudes(jde float64) (lon [9]float64, sunDist, moonDist float64) {
	Δψ, _ := nutation.Nutation(jde)
	tc := (jde - base.J2000) / 36525 // Julian centuries

	l0, b0, r0 := earth.position(jde)
	sunLon, _ := toFK5(l0+math.Pi, -b0, tc)
	lon[zodiac.Sun] = sunLon + Δψ.Rad() - 20.4898*arcsec/r0 // with aberration

	for p, s := range series {
		lon[p] = planet(s, jde, l0, b0, r0, tc, sunLon) + Δψ.Rad()
	}

	λ, _, Δ := moonposition.Position(jde)
	lon[zodiac.Moon] = λ.Rad() + Δψ.Rad()
	lon[zodiac.NorthNode] = moonposition.Node(jde).Rad()
	lon[zodiac.SouthNode] = lon[zodiac.NorthNode] + math.Pi

	for i := range lon {
		lon[i] = norm(lon[i])
	}
	return lon, r0, Δ
}

// planet returns a planet's geocentric longitude at jde, corrected for
// light-time, FK5 and aberration, but not nutation (Meeus, chapter 33). The
// Earth is at l0, b0, r0; tc is Julian centuries from J2000; the Sun's
// longitude is sunLon.
func planet(s *vsop87, jde, l0, b0, r0, tc, sunLon float64) float64 {
	var x, y, z float64
	τ := 0.0
	for range 3 {
		l, b, r := s.position(jde - τ)
		x = r*math.Cos(b)*math.Cos(l) - r0*math.Cos(b0)*math.Cos(l0)
		y = r*math.Cos(b)*math.Sin(l) - r0*math.Cos(b0)*math.Sin(l0)
		z = r*math.Sin(b) - r0*math.Sin(b0)
		τ = 0.0057755183 * math.Sqrt(x*x+y*y+z*z) // light-time, days
	}
	λ := math.Atan2(y, x)
	β := math.Atan2(z, math.Hypot(x, y))
	λ, β = toFK5(λ, β, tc)
	// Annual aberration, without the small terms in the Earth's
	// eccentricity (Meeus 23.2).
	κ := 20.49552 * arcsec
	return λ - κ*math.Cos(sunLon-λ)/math.Cos(β)
}

// toFK5 converts VSOP87 coordinates to the FK5 system (Meeus 32.3), tc
// Julian centuries from J2000.
func toFK5(λ, β, tc float64) (float64, float64) {
	λp := λ - (1.397*tc+0.00031*tc*tc)*math.Pi/180
	dλ := -0.09033 + 0.03916*(math.Cos(λp)+math.Sin(λp))*math.Tan(β)
	dβ := 0.03916 * (math.Cos(λp) - math.Sin(λp))
	return λ + dλ*arcsec, β + dβ*arcsec
}

// norm reduces an angle to [0, 2π).
func norm(a float64) float64 {
	a = math.Mod(a, 2*math.Pi)
	if a < 0 {
		a += 2 * math.Pi
	}
	return a
}
