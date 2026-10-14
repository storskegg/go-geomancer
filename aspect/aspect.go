// Package aspect works out the geomantic aspects to the significator: the
// houses that stand in Opposition, Square, Trine or Sextile to the house of
// the question. The rules are the same in every system.
package aspect

import (
	"errors"
	"fmt"
)

// Aspect is a relationship between the significator's house and another
// house, found by counting around the houses.
//
// Counting includes the significator's house, so a Sextile, "count three",
// is two houses away, a Square, "count four", is three away, and a Trine,
// "count five", is four away. The Opposition is the house directly
// opposite, six away.
type Aspect uint8

const (
	Opposition Aspect = iota
	Square
	Trine
	Sextile
)

var aspectInfo = []struct {
	name, symbol string
	distance     int
}{
	Opposition: {"Opposition", "☍", 6},
	Square:     {"Square", "□", 3},
	Trine:      {"Trine", "△", 4},
	Sextile:    {"Sextile", "⚹", 2},
}

// ErrInvalidHouse is returned for a house number outside 1–12.
var ErrInvalidHouse = errors.New("aspect: invalid house")

// All returns every aspect: Opposition, Square, Trine, Sextile.
func All() []Aspect { return []Aspect{Opposition, Square, Trine, Sextile} }

// Valid reports whether a is a defined aspect.
func (a Aspect) Valid() bool { return int(a) < len(aspectInfo) }

// Distance returns how many houses away from the significator the aspect
// falls.
func (a Aspect) Distance() int { return aspectInfo[a].distance }

// Symbol returns the aspect's astrological symbol.
func (a Aspect) Symbol() string { return aspectInfo[a].symbol }

// Stressful reports whether the aspect is stressful (Opposition and Square)
// rather than easing the situation (Trine and Sextile; the Sextile is mildly
// benevolent).
func (a Aspect) Stressful() bool { return a == Opposition || a == Square }

func (a Aspect) String() string {
	if a.Valid() {
		return aspectInfo[a].name
	}
	return fmt.Sprintf("Aspect(%d)", a)
}

func (a Aspect) MarshalText() ([]byte, error) {
	if !a.Valid() {
		return nil, fmt.Errorf("aspect: invalid aspect %d", a)
	}
	return []byte(a.String()), nil
}

func (a *Aspect) UnmarshalText(b []byte) error {
	for _, v := range All() {
		if v.String() == string(b) {
			*a = v
			return nil
		}
	}
	return fmt.Errorf("aspect: unknown aspect %q", b)
}

// Side is the side of the significator an aspect falls on. Each aspect but
// the Opposition falls on both sides, and of the pair the Dexter house is
// the more powerful.
type Side uint8

const (
	// Neither is the Opposition's side: it falls on both.
	Neither Side = iota
	// Dexter is clockwise around the Square, counting back through the
	// house numbers.
	Dexter
	// Sinister is counterclockwise, counting on through the house numbers.
	Sinister
)

var sideNames = []string{"", "Dexter", "Sinister"}

func (s Side) String() string {
	if int(s) < len(sideNames) {
		return sideNames[s]
	}
	return fmt.Sprintf("Side(%d)", s)
}

func (s Side) MarshalText() ([]byte, error) { return []byte(s.String()), nil }

func (s *Side) UnmarshalText(b []byte) error {
	for i, n := range sideNames {
		if n == string(b) {
			*s = Side(i)
			return nil
		}
	}
	return fmt.Errorf("aspect: unknown side %q", b)
}

// House is a house aspecting the significator.
type House struct {
	House  int    `json:"house"`
	Aspect Aspect `json:"aspect"`
	Side   Side   `json:"side,omitempty"` // omitted for the Opposition
}

// Houses returns the houses aspecting the significator's house: the
// Opposition, then the Dexter and Sinister Square, Trine and Sextile.
func Houses(significator int) ([]House, error) {
	if significator < 1 || significator > 12 {
		return nil, fmt.Errorf("%w: %d, want 1–12", ErrInvalidHouse, significator)
	}
	houses := []House{{House: houseAt(significator, Opposition.Distance()), Aspect: Opposition}}
	for _, a := range []Aspect{Square, Trine, Sextile} {
		houses = append(houses,
			House{House: houseAt(significator, -a.Distance()), Aspect: a, Side: Dexter},
			House{House: houseAt(significator, a.Distance()), Aspect: a, Side: Sinister},
		)
	}
	return houses, nil
}

// houseAt returns the house n places after house h, wrapping around 1–12.
func houseAt(h, n int) int { return ((h-1+n)%12+12)%12 + 1 }
