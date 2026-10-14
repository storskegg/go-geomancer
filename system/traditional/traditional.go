// Package traditional provides the Traditional system of geomancy and
// registers it with package system.
package traditional

import (
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/zodiac"
)

// System is the Traditional system.
var System system.System = &system.Mapping{
	Slug:  "traditional",
	Title: "Traditional",
	SegmentElements: [4]zodiac.Element{
		figure.Head: zodiac.Fire,
		figure.Neck: zodiac.Air,
		figure.Body: zodiac.Water,
		figure.Feet: zodiac.Earth,
	},
	Houses: [12]int{
		position.M1: 1, position.M2: 2, position.M3: 3, position.M4: 4,
		position.D1: 5, position.D2: 6, position.D3: 7, position.D4: 8,
		position.N1: 9, position.N2: 10, position.N3: 11, position.N4: 12,
	},
	Signs:    signs,
	Elements: elements,
}

// signs gives each figure its zodiacal sign.
//
// TODO: confirm which table "traditional" uses and fill this in. Until then
// the Astrological Square has no signs under this system. One candidate,
// recalled from Gerard of Cremona and NOT verified against a source:
//
//	Acquisitio: Aries        Puella: Libra
//	Laetitia: Taurus         Tristitia: Scorpio
//	Amissio: Taurus          Cauda Draconis: Sagittarius
//	Fortuna Minor: Taurus    Populus: Capricorn
//	Puer: Gemini             Fortuna Major: Aquarius
//	Rubeus: Gemini           Carcer: Pisces
//	Albus: Cancer            Conjunctio: Virgo or Pisces (unsure)
//	Via: Leo
//	Caput Draconis: Virgo
var signs = map[figure.Figure]zodiac.Sign{}

// elements gives each figure its element.
//
// TODO: fill in from the preferred source.
var elements = map[figure.Figure]zodiac.Element{}

func init() { system.Register(System) }
