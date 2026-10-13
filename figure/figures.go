package figure

import "github.com/storskegg/go-geomancer/zodiac"

// The sixteen figures. Comments give the dot pattern top to bottom.
const (
	Populus       Figure = 0b0000 // 2222
	Via           Figure = 0b1111 // 1111
	Albus         Figure = 0b0010 // 2212
	Conjunctio    Figure = 0b0110 // 2112
	Puella        Figure = 0b1011 // 1211
	Amissio       Figure = 0b1010 // 1212
	FortunaMajor  Figure = 0b0011 // 2211
	FortunaMinor  Figure = 0b1100 // 1122
	Puer          Figure = 0b1101 // 1121
	Rubeus        Figure = 0b0100 // 2122
	Acquisitio    Figure = 0b0101 // 2121
	Laetitia      Figure = 0b1000 // 1222
	Carcer        Figure = 0b1001 // 1221
	Tristitia     Figure = 0b0001 // 2221
	CaputDraconis Figure = 0b0111 // 2111
	CaudaDraconis Figure = 0b1110 // 1112
)

type info struct {
	name    string
	planet  zodiac.Planet
	aliases []string
}

// figures is indexed by Figure value.
var figures = [16]info{
	Populus:       {name: "Populus", planet: zodiac.Moon},
	Via:           {name: "Via", planet: zodiac.Moon},
	Albus:         {name: "Albus", planet: zodiac.Mercury},
	Conjunctio:    {name: "Conjunctio", planet: zodiac.Mercury, aliases: []string{"Coniunctio"}},
	Puella:        {name: "Puella", planet: zodiac.Venus},
	Amissio:       {name: "Amissio", planet: zodiac.Venus},
	FortunaMajor:  {name: "Fortuna Major", planet: zodiac.Sun, aliases: []string{"Fortuna Maior"}},
	FortunaMinor:  {name: "Fortuna Minor", planet: zodiac.Sun},
	Puer:          {name: "Puer", planet: zodiac.Mars},
	Rubeus:        {name: "Rubeus", planet: zodiac.Mars},
	Acquisitio:    {name: "Acquisitio", planet: zodiac.Jupiter},
	Laetitia:      {name: "Laetitia", planet: zodiac.Jupiter, aliases: []string{"Letitia"}},
	Carcer:        {name: "Carcer", planet: zodiac.Saturn},
	Tristitia:     {name: "Tristitia", planet: zodiac.Saturn},
	CaputDraconis: {name: "Caput Draconis", planet: zodiac.NorthNode},
	CaudaDraconis: {name: "Cauda Draconis", planet: zodiac.SouthNode},
}

// All returns the sixteen figures, grouped in pairs by planet.
func All() [16]Figure {
	return [16]Figure{
		Via, Populus,
		Albus, Conjunctio,
		Puella, Amissio,
		FortunaMajor, FortunaMinor,
		Puer, Rubeus,
		Acquisitio, Laetitia,
		Carcer, Tristitia,
		CaputDraconis, CaudaDraconis,
	}
}
