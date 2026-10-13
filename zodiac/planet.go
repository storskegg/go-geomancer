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

// Planets returns all planets in order.
func Planets() []Planet {
	return []Planet{Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn, NorthNode, SouthNode}
}

// ParsePlanet parses a planet name, ignoring case and spacing.
func ParsePlanet(s string) (Planet, error) { return parse[Planet](planetLabels, "planet", s) }

// Valid reports whether p is a defined planet.
func (p Planet) Valid() bool { return int(p) < len(planetLabels) }

func (p Planet) String() string { return label(planetLabels, "Planet", p) }

func (p Planet) MarshalText() ([]byte, error) { return marshal(planetLabels, "planet", p) }

func (p *Planet) UnmarshalText(b []byte) error {
	v, err := ParsePlanet(string(b))
	if err != nil {
		return err
	}
	*p = v
	return nil
}
