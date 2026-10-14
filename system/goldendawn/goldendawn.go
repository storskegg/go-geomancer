// Package goldendawn provides the Golden Dawn system of geomancy and
// registers it with package system.
package goldendawn

import (
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/zodiac"
)

// System is the Golden Dawn system.
var System system.System = &system.Mapping{
	Slug:  "goldendawn",
	Title: "Golden Dawn",
	SegmentElements: [4]zodiac.Element{
		figure.Head: zodiac.Fire,
		figure.Neck: zodiac.Water,
		figure.Body: zodiac.Air,
		figure.Feet: zodiac.Earth,
	},
	Houses: [12]int{
		position.M1: 10, position.M2: 1, position.M3: 4, position.M4: 7,
		position.D1: 11, position.D2: 2, position.D3: 5, position.D4: 8,
		position.N1: 12, position.N2: 3, position.N3: 6, position.N4: 9,
	},
	Signs:    signs,
	Elements: elements,
}

// signs gives each figure the sign ruled by its planet, one figure per sign
// with the Moon's and the Sun's pairs sharing Cancer and Leo.
//
// TODO: verify against the preferred Golden Dawn source. Caput Draconis and
// Cauda Draconis are commonly given no sign; when one falls in the 1st House
// the Astrological Square has no signs.
var signs = map[figure.Figure]zodiac.Sign{
	figure.Puer:         zodiac.Aries,
	figure.Amissio:      zodiac.Taurus,
	figure.Albus:        zodiac.Gemini,
	figure.Populus:      zodiac.Cancer,
	figure.Via:          zodiac.Cancer,
	figure.FortunaMajor: zodiac.Leo,
	figure.FortunaMinor: zodiac.Leo,
	figure.Conjunctio:   zodiac.Virgo,
	figure.Puella:       zodiac.Libra,
	figure.Rubeus:       zodiac.Scorpio,
	figure.Acquisitio:   zodiac.Sagittarius,
	figure.Carcer:       zodiac.Capricorn,
	figure.Tristitia:    zodiac.Aquarius,
	figure.Laetitia:     zodiac.Pisces,
}

// elements gives each figure the element of its sign, with Earth for Caput
// Draconis and Fire for Cauda Draconis.
//
// TODO: verify against the preferred Golden Dawn source. These were derived
// from the signs above rather than taken from a published table.
var elements = map[figure.Figure]zodiac.Element{
	figure.Puer:          zodiac.Fire,
	figure.Amissio:       zodiac.Earth,
	figure.Albus:         zodiac.Air,
	figure.Populus:       zodiac.Water,
	figure.Via:           zodiac.Water,
	figure.FortunaMajor:  zodiac.Fire,
	figure.FortunaMinor:  zodiac.Fire,
	figure.Conjunctio:    zodiac.Earth,
	figure.Puella:        zodiac.Air,
	figure.Rubeus:        zodiac.Water,
	figure.Acquisitio:    zodiac.Fire,
	figure.Carcer:        zodiac.Earth,
	figure.Tristitia:     zodiac.Air,
	figure.Laetitia:      zodiac.Water,
	figure.CaputDraconis: zodiac.Earth,
	figure.CaudaDraconis: zodiac.Fire,
}

func init() { system.Register(System) }
