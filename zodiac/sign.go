package zodiac

// Sign is one of the twelve zodiacal signs, in zodiacal order from Aries.
type Sign uint8

const (
	Aries Sign = iota
	Taurus
	Gemini
	Cancer
	Leo
	Virgo
	Libra
	Scorpio
	Sagittarius
	Capricorn
	Aquarius
	Pisces
)

var signLabels = []string{
	"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo",
	"Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces",
}

// Signs returns all signs in zodiacal order.
func Signs() []Sign {
	return []Sign{
		Aries, Taurus, Gemini, Cancer, Leo, Virgo,
		Libra, Scorpio, Sagittarius, Capricorn, Aquarius, Pisces,
	}
}

// ParseSign parses a sign name, ignoring case.
func ParseSign(s string) (Sign, error) { return parse[Sign](signLabels, "sign", s) }

// Valid reports whether s is a defined sign.
func (s Sign) Valid() bool { return int(s) < len(signLabels) }

// Add returns the sign n places after s in zodiacal order, wrapping past
// Pisces. Negative n counts backwards. If the 1st House holds Cancer, House k
// holds Cancer.Add(k-1).
func (s Sign) Add(n int) Sign {
	const count = 12
	return Sign(((int(s)+n)%count + count) % count)
}

// Next returns the following sign in zodiacal order.
func (s Sign) Next() Sign { return s.Add(1) }

func (s Sign) String() string { return label(signLabels, "Sign", s) }

func (s Sign) MarshalText() ([]byte, error) { return marshal(signLabels, "sign", s) }

func (s *Sign) UnmarshalText(b []byte) error {
	v, err := ParseSign(string(b))
	if err != nil {
		return err
	}
	*s = v
	return nil
}
