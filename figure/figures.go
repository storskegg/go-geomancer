package figure

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
	aliases []string
}

// figures is indexed by Figure value.
var figures = [16]info{
	Populus:       {name: "Populus"},
	Via:           {name: "Via"},
	Albus:         {name: "Albus"},
	Conjunctio:    {name: "Conjunctio", aliases: []string{"Coniunctio"}},
	Puella:        {name: "Puella"},
	Amissio:       {name: "Amissio"},
	FortunaMajor:  {name: "Fortuna Major", aliases: []string{"Fortuna Maior"}},
	FortunaMinor:  {name: "Fortuna Minor"},
	Puer:          {name: "Puer"},
	Rubeus:        {name: "Rubeus"},
	Acquisitio:    {name: "Acquisitio"},
	Laetitia:      {name: "Laetitia", aliases: []string{"Letitia"}},
	Carcer:        {name: "Carcer"},
	Tristitia:     {name: "Tristitia"},
	CaputDraconis: {name: "Caput Draconis"},
	CaudaDraconis: {name: "Cauda Draconis"},
}

// All returns the sixteen figures, in traditional order: paired by planet,
// the Moon's first.
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
