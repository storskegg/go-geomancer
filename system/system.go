// Package system defines geomantic Systems: the correspondences that differ
// between traditions, such as the elements of the segments, the houses the
// principal figures are placed in, and the signs of the figures.
//
// Concrete systems live in subpackages and register themselves on import:
//
//	import _ "github.com/storskegg/go-geomancer/system/traditional"
package system

import (
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/zodiac"
)

// System is a set of geomantic correspondences.
type System interface {
	// ID is a stable lowercase identifier, such as "traditional".
	ID() string
	// Name is the human-readable name, such as "Traditional".
	Name() string
	// SegmentElement returns the element of a figure segment.
	SegmentElement(figure.Segment) zodiac.Element
	// House returns the house (1–12) the figure at p is placed in. ok is
	// false for positions that aren't placed in a house.
	House(p position.Position) (house int, ok bool)
	// Sign returns the figure's zodiacal sign. ok is false if the system
	// doesn't give the figure a sign.
	Sign(figure.Figure) (sign zodiac.Sign, ok bool)
	// Element returns the figure's element. ok is false if the system
	// doesn't give the figure an element.
	Element(figure.Figure) (element zodiac.Element, ok bool)
}

// Mapping is a System defined by lookup tables.
type Mapping struct {
	Slug  string
	Title string
	// SegmentElements is indexed by figure.Segment.
	SegmentElements [4]zodiac.Element
	// Houses gives the house (1–12) of each principal position, indexed
	// position.M1 through position.N4.
	Houses [12]int
	// Signs and Elements give the figures' correspondences. Figures missing
	// from a map are unmapped.
	Signs    map[figure.Figure]zodiac.Sign
	Elements map[figure.Figure]zodiac.Element
}

func (m Mapping) ID() string   { return m.Slug }
func (m Mapping) Name() string { return m.Title }

// SegmentElement panics if s is not a valid segment.
func (m Mapping) SegmentElement(s figure.Segment) zodiac.Element { return m.SegmentElements[s] }

func (m Mapping) House(p position.Position) (int, bool) {
	if !p.Principal() {
		return 0, false
	}
	return m.Houses[p], true
}

func (m Mapping) Sign(f figure.Figure) (zodiac.Sign, bool) {
	s, ok := m.Signs[f]
	return s, ok
}

func (m Mapping) Element(f figure.Figure) (zodiac.Element, bool) {
	e, ok := m.Elements[f]
	return e, ok
}
