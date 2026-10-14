// Package figure implements geomantic figures (tetragrams): four segments,
// each Odd or Even, and the arithmetic used to derive one figure from others.
package figure

import (
	"fmt"
	"strings"
)

// Figure is a geomantic figure stored as a 4-bit value. Head is the most
// significant bit and a set bit means Odd, so the binary literal reads top to
// bottom: Amissio (odd, even, odd, even) is 0b1010.
//
// With this encoding, adding figures is XOR.
type Figure uint8

const mask Figure = 0b1111

func bit(s Segment) Figure { return 1 << (3 - s) }

// New builds a figure from its segments, top to bottom.
func New(head, neck, body, feet Parity) Figure {
	return FromSegments([4]Parity{head, neck, body, feet})
}

// FromSegments builds a figure from its segments, top to bottom.
func FromSegments(p [4]Parity) Figure {
	var f Figure
	for s, v := range p {
		if v == Odd {
			f |= bit(Segment(s))
		}
	}
	return f
}

// Valid reports whether f is one of the 16 figures.
func (f Figure) Valid() bool { return f&^mask == 0 }

// Segment returns the parity of segment s.
func (f Figure) Segment(s Segment) Parity {
	if f&bit(s) != 0 {
		return Odd
	}
	return Even
}

// Segments returns the parity of each segment, top to bottom.
func (f Figure) Segments() [4]Parity {
	var p [4]Parity
	for _, s := range Segments() {
		p[s] = f.Segment(s)
	}
	return p
}

// Add returns the sum of f and g: each segment of the result is the parity
// sum of the corresponding segments.
func (f Figure) Add(g Figure) Figure { return (f ^ g) & mask }

// OddDots returns the number of dots in odd segments (one per segment).
func (f Figure) OddDots() int {
	n := 0
	for _, p := range f.Segments() {
		if p == Odd {
			n++
		}
	}
	return n
}

// EvenDots returns the number of dots in even segments (two per segment).
func (f Figure) EvenDots() int { return (4 - f.OddDots()) * 2 }

// Dots returns the total number of dots in the figure.
func (f Figure) Dots() int { return f.OddDots() + f.EvenDots() }

// Name returns the figure's Latin name.
func (f Figure) Name() string {
	if !f.Valid() {
		return fmt.Sprintf("Figure(%d)", uint8(f))
	}
	return figures[f].name
}

// Pattern returns the figure as dot counts top to bottom, e.g. "1212" for
// Amissio.
func (f Figure) Pattern() string {
	var b strings.Builder
	for _, p := range f.Segments() {
		b.WriteByte(byte('0' + p.Dots()))
	}
	return b.String()
}

func (f Figure) String() string { return f.Name() }

// MarshalText encodes the figure as its name.
func (f Figure) MarshalText() ([]byte, error) {
	if !f.Valid() {
		return nil, fmt.Errorf("%w: %d", ErrInvalid, uint8(f))
	}
	return []byte(f.Name()), nil
}

// UnmarshalText decodes any form accepted by Parse.
func (f *Figure) UnmarshalText(b []byte) error {
	v, err := Parse(string(b))
	if err != nil {
		return err
	}
	*f = v
	return nil
}
