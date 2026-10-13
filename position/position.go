// Package position names the slots of a geomantic chart (M1…N4, the
// Witnesses, the Judge and the Judgement). It is a leaf package so that both
// chart and system can refer to positions without importing each other.
package position

import (
	"errors"
	"fmt"
	"strings"
)

// Position is a slot in the Shield that holds a figure.
type Position uint8

const (
	M1 Position = iota
	M2
	M3
	M4
	D1
	D2
	D3
	D4
	N1
	N2
	N3
	N4
	WR // Right Witness
	WL // Left Witness
	J  // Judge
	JM // Judgement
)

var labels = []string{
	"M1", "M2", "M3", "M4",
	"D1", "D2", "D3", "D4",
	"N1", "N2", "N3", "N4",
	"WR", "WL", "J", "JM",
}

// ErrUnknown is returned when parsing a string that isn't a position code.
var ErrUnknown = errors.New("position: unknown position")

// All returns every position in chart order.
func All() []Position {
	ps := make([]Position, len(labels))
	for i := range ps {
		ps[i] = Position(i)
	}
	return ps
}

// Principals returns the twelve principal positions, M1 through N4, which
// are placed in the houses of the Astrological Square.
func Principals() [12]Position {
	return [12]Position{M1, M2, M3, M4, D1, D2, D3, D4, N1, N2, N3, N4}
}

// Parse reads a position code such as "M1" or "wr".
func Parse(s string) (Position, error) {
	s = strings.TrimSpace(s)
	for i, l := range labels {
		if strings.EqualFold(l, s) {
			return Position(i), nil
		}
	}
	return 0, fmt.Errorf("%w: %q", ErrUnknown, s)
}

// Valid reports whether p is a defined position.
func (p Position) Valid() bool { return int(p) < len(labels) }

// Principal reports whether p is one of the twelve principal positions.
func (p Position) Principal() bool { return p <= N4 }

func (p Position) String() string {
	if p.Valid() {
		return labels[p]
	}
	return fmt.Sprintf("Position(%d)", p)
}

func (p Position) MarshalText() ([]byte, error) {
	if !p.Valid() {
		return nil, fmt.Errorf("%w: %d", ErrUnknown, p)
	}
	return []byte(labels[p]), nil
}

func (p *Position) UnmarshalText(b []byte) error {
	v, err := Parse(string(b))
	if err != nil {
		return err
	}
	*p = v
	return nil
}
