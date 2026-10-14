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

// signs gives each figure its zodiacal sign.
//
// TODO: fill in. Until then the Astrological Square has no signs under this
// system.
var signs = map[figure.Figure]zodiac.Sign{}

// elements gives each figure its element.
//
// TODO: fill in.
var elements = map[figure.Figure]zodiac.Element{}

func init() { system.Register(System) }
