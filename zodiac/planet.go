package zodiac

// Planet is one of the seven classical planets or a lunar node.
type Planet uint8

const (
	Sun Planet = iota
	Moon
	Mercury
	Venus
	Mars
	Jupiter
	Saturn
	NorthNode
	SouthNode
)

var planetLabels = []string{
	"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "North Node", "South Node",
}

// Unicode symbols, and conventional two-letter abbreviations for terminals
// whose fonts lack them.
var (
	planetSymbols       = []string{"☉", "☽", "☿", "♀", "♂", "♃", "♄", "☊", "☋"}
	planetAbbreviations = []string{"Su", "Mo", "Me", "Ve", "Ma", "Ju", "Sa", "NN", "SN"}
)

// Planets returns all planets in order.
func Planets() []Planet {
	return []Planet{Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn, NorthNode, SouthNode}
}

// ParsePlanet parses a planet name, ignoring case and spacing.
func ParsePlanet(s string) (Planet, error) { return parse[Planet](planetLabels, "planet", s) }

// Valid reports whether p is a defined planet.
func (p Planet) Valid() bool { return int(p) < len(planetLabels) }

func (p Planet) String() string { return label(planetLabels, "Planet", p) }

// Symbol returns the planet's symbol, such as ☉ for the Sun.
func (p Planet) Symbol() string { return label(planetSymbols, "Planet", p) }

// Abbreviation returns the planet's two-letter abbreviation, such as "Su".
func (p Planet) Abbreviation() string { return label(planetAbbreviations, "Planet", p) }

func (p Planet) MarshalText() ([]byte, error) { return marshal(planetLabels, "planet", p) }

func (p *Planet) UnmarshalText(b []byte) error {
	v, err := ParsePlanet(string(b))
	if err != nil {
		return err
	}
	*p = v
	return nil
}
