package sky

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math"
	"slices"

	"github.com/storskegg/go-geomancer/zodiac"
)

// Dignity is a planet's essential dignity or debility by sign, under the
// traditional rulerships.
type Dignity uint8

const (
	Domicile Dignity = iota // in a sign it rules
	Exaltation
	Detriment // opposite a sign it rules
	Fall      // opposite its exaltation
	Peregrine // none of these
)

var dignityLabels = []string{"Domicile", "Exaltation", "Detriment", "Fall", "Peregrine"}

func (d Dignity) String() string                { return label(dignityLabels, "Dignity", d) }
func (d Dignity) MarshalText() ([]byte, error)  { return marshal(dignityLabels, "dignity", d) }
func (d *Dignity) UnmarshalText(b []byte) error { return unmarshal(dignityLabels, "dignity", b, d) }

// rulers gives each sign's traditional ruler, Aries first.
var rulers = [12]zodiac.Planet{
	zodiac.Mars, zodiac.Venus, zodiac.Mercury, zodiac.Moon, zodiac.Sun, zodiac.Mercury,
	zodiac.Venus, zodiac.Mars, zodiac.Jupiter, zodiac.Saturn, zodiac.Saturn, zodiac.Jupiter,
}

// exaltations gives each planet's sign of exaltation.
var exaltations = map[zodiac.Planet]zodiac.Sign{
	zodiac.Sun: zodiac.Aries, zodiac.Moon: zodiac.Taurus, zodiac.Mercury: zodiac.Virgo, zodiac.Venus: zodiac.Pisces,
	zodiac.Mars: zodiac.Capricorn, zodiac.Jupiter: zodiac.Cancer, zodiac.Saturn: zodiac.Libra,
}

// dignities returns p's dignities and debilities in sign s, or Peregrine.
// Mercury in Virgo is both in its domicile and exalted.
func dignities(p zodiac.Planet, s zodiac.Sign) []Dignity {
	var ds []Dignity
	if rulers[s] == p {
		ds = append(ds, Domicile)
	}
	if exaltations[p] == s {
		ds = append(ds, Exaltation)
	}
	if rulers[s.Add(6)] == p {
		ds = append(ds, Detriment)
	}
	if exaltations[p].Add(6) == s {
		ds = append(ds, Fall)
	}
	if len(ds) == 0 {
		ds = append(ds, Peregrine)
	}
	return ds
}

// Solar is a planet's closeness to the Sun in longitude.
type Solar uint8

const (
	NotNearSun    Solar = iota
	Cazimi              // within 17′: "in the heart of the Sun", strengthened
	Combust             // within 8°30′: burnt up, weakened
	UnderTheBeams       // within 17°: hidden, somewhat weakened
)

var solarLabels = []string{"", "Cazimi", "Combust", "Under the beams"}

func (s Solar) String() string                { return label(solarLabels, "Solar", s) }
func (s Solar) MarshalText() ([]byte, error)  { return marshal(solarLabels, "solar", s) }
func (s *Solar) UnmarshalText(b []byte) error { return unmarshal(solarLabels, "solar", b, s) }

// solar classifies a distance from the Sun, in radians.
func solar(d float64) Solar {
	switch deg := d * 180 / math.Pi; {
	case deg <= 17.0/60:
		return Cazimi
	case deg <= 8.5:
		return Combust
	case deg <= 17:
		return UnderTheBeams
	}
	return NotNearSun
}

// Kind is one of the Ptolemaic aspects.
type Kind uint8

const (
	Conjunction Kind = iota
	Sextile
	Square
	Trine
	Opposition
)

var kindInfo = []struct {
	name, symbol string
	angle        float64 // degrees
}{
	Conjunction: {"Conjunction", "☌", 0},
	Sextile:     {"Sextile", "⚹", 60},
	Square:      {"Square", "□", 90},
	Trine:       {"Trine", "△", 120},
	Opposition:  {"Opposition", "☍", 180},
}

func (k Kind) String() string {
	if int(k) < len(kindInfo) {
		return kindInfo[k].name
	}
	return fmt.Sprintf("Kind(%d)", k)
}

// Symbol returns the aspect's symbol, such as □.
func (k Kind) Symbol() string {
	if int(k) < len(kindInfo) {
		return kindInfo[k].symbol
	}
	return "?"
}

func (k Kind) MarshalText() ([]byte, error) {
	if int(k) >= len(kindInfo) {
		return nil, fmt.Errorf("sky: unknown aspect %d", k)
	}
	return []byte(kindInfo[k].name), nil
}

func (k *Kind) UnmarshalText(b []byte) error {
	for i, info := range kindInfo {
		if info.name == string(b) {
			*k = Kind(i)
			return nil
		}
	}
	return fmt.Errorf("sky: unknown aspect %q", b)
}

// Aspect is an aspect between two bodies.
type Aspect struct {
	A, B zodiac.Planet `json:"-"`
	Kind Kind          `json:"kind"`
	// Orb is how far, in degrees, the bodies are from the exact aspect.
	Orb float64 `json:"orb"`
	// Applying is set when the aspect is becoming exact, and clear when
	// it's separating.
	Applying bool `json:"applying"`
}

// MarshalJSON includes the bodies by name.
func (a Aspect) MarshalJSON() ([]byte, error) {
	type plain Aspect
	return json.Marshal(struct {
		Bodies [2]zodiac.Planet `json:"bodies"`
		plain
	}{[2]zodiac.Planet{a.A, a.B}, plain(a)})
}

// UnmarshalJSON reads what MarshalJSON writes.
func (a *Aspect) UnmarshalJSON(b []byte) error {
	type plain Aspect
	var v struct {
		Bodies [2]zodiac.Planet `json:"bodies"`
		plain
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*a = Aspect(v.plain)
	a.A, a.B = v.Bodies[0], v.Bodies[1]
	return nil
}

// moieties are half the planets' orbs, in degrees, after Lilly. Two planets
// are in aspect within the sum of their moieties.
var moieties = map[zodiac.Planet]float64{
	zodiac.Sun: 7.5, zodiac.Moon: 6, zodiac.Mercury: 3.5, zodiac.Venus: 3.5,
	zodiac.Mars: 3.75, zodiac.Jupiter: 4.5, zodiac.Saturn: 4.5,
}

var planets = []zodiac.Planet{zodiac.Sun, zodiac.Moon, zodiac.Mercury, zodiac.Venus, zodiac.Mars, zodiac.Jupiter, zodiac.Saturn}

// aspects finds the aspects among the planets at longitudes now, and their
// conjunctions with the nodes, using the longitudes a little later to tell
// applying from separating. Closest first.
func aspects(now, later [9]float64) []Aspect {
	var out []Aspect
	try := func(a, b zodiac.Planet, kinds []Kind, orb float64) {
		for _, k := range kinds {
			exact := kindInfo[k].angle * math.Pi / 180
			o := math.Abs(angle(now[a], now[b]) - exact)
			if o*180/math.Pi <= orb {
				o2 := math.Abs(angle(later[a], later[b]) - exact)
				out = append(out, Aspect{A: a, B: b, Kind: k, Orb: o * 180 / math.Pi, Applying: o2 < o})
			}
		}
	}
	all := []Kind{Conjunction, Sextile, Square, Trine, Opposition}
	for i, a := range planets {
		for _, b := range planets[i+1:] {
			try(a, b, all, moieties[a]+moieties[b])
		}
		for _, n := range []zodiac.Planet{zodiac.NorthNode, zodiac.SouthNode} {
			try(a, n, []Kind{Conjunction}, moieties[a])
		}
	}
	slices.SortStableFunc(out, func(x, y Aspect) int { return cmp.Compare(x.Orb, y.Orb) })
	return out
}
