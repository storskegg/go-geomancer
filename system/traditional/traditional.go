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
// TODO: confirm the traditional table and update this. Until then it copies
// the Golden Dawn table (see system/goldendawn), which follows planetary
// rulership. One candidate traditional table, recalled from Gerard of
// Cremona and NOT verified against a source:
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

// elements gives each figure its element.
//
// TODO: confirm the traditional table and update this. Until then it copies
// the Golden Dawn table (see system/goldendawn).
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
