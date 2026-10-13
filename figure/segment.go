package figure

import "fmt"

// Segment is one of the four rows of a figure, named for the part of the body
// it corresponds to, top to bottom.
type Segment uint8

const (
	Head Segment = iota
	Neck
	Body
	Feet
)

var segmentLabels = [...]string{"Head", "Neck", "Body", "Feet"}

// Segments returns the four segments, top to bottom.
func Segments() [4]Segment { return [4]Segment{Head, Neck, Body, Feet} }

func (s Segment) String() string {
	if int(s) < len(segmentLabels) {
		return segmentLabels[s]
	}
	return fmt.Sprintf("Segment(%d)", s)
}

// Parity is the value of a single segment: Odd (one dot, active) or Even (two
// dots, passive).
type Parity uint8

const (
	Even Parity = iota
	Odd
)

// Add returns the parity of the sum of p and q: Odd + Even is Odd, while
// Odd + Odd and Even + Even are Even.
func (p Parity) Add(q Parity) Parity { return p ^ q }

// Dots returns the number of dots drawn for the segment: 1 for Odd, 2 for Even.
func (p Parity) Dots() int {
	if p == Odd {
		return 1
	}
	return 2
}

// Active reports whether the segment is active, i.e. Odd.
func (p Parity) Active() bool { return p == Odd }

func (p Parity) String() string {
	if p == Odd {
		return "odd"
	}
	return "even"
}
