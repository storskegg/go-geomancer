package zodiac

// Element is one of the four classical elements.
type Element uint8

const (
	Fire Element = iota
	Air
	Water
	Earth
)

var elementLabels = []string{"Fire", "Air", "Water", "Earth"}

// Alchemical symbols. Many terminal fonts lack these.
var elementSymbols = []string{"🜂", "🜁", "🜄", "🜃"}

// Elements returns all elements in order.
func Elements() []Element { return []Element{Fire, Air, Water, Earth} }

// ParseElement parses an element name, ignoring case.
func ParseElement(s string) (Element, error) { return parse[Element](elementLabels, "element", s) }

// Valid reports whether e is a defined element.
func (e Element) Valid() bool { return int(e) < len(elementLabels) }

func (e Element) String() string { return label(elementLabels, "Element", e) }

// Symbol returns the element's alchemical symbol: 🜂 Fire, 🜁 Air, 🜄 Water,
// 🜃 Earth.
func (e Element) Symbol() string { return label(elementSymbols, "Element", e) }

func (e Element) MarshalText() ([]byte, error) { return marshal(elementLabels, "element", e) }

func (e *Element) UnmarshalText(b []byte) error {
	v, err := ParseElement(string(b))
	if err != nil {
		return err
	}
	*e = v
	return nil
}
