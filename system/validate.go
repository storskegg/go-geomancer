package system

import (
	"errors"
	"fmt"
	"strings"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/zodiac"
)

// ErrInvalid is returned by Validate for a system with inconsistent
// mappings.
var ErrInvalid = errors.New("system: invalid system")

// Validate checks that a system is internally consistent:
//
//   - its ID is non-empty and lowercase
//   - each segment has a distinct, valid element
//   - the twelve principal positions are placed one per house, 1–12, and no
//     other position is placed in a house
//   - any signs and elements it gives figures are valid
//
// Sign and element tables may be partial or empty.
func Validate(s System) error {
	var problems []string
	add := func(format string, args ...any) { problems = append(problems, fmt.Sprintf(format, args...)) }

	if id := s.ID(); id == "" || id != strings.ToLower(id) || strings.ContainsAny(id, " \t") {
		add("ID %q must be non-empty, lowercase and without spaces", id)
	}

	elements := map[zodiac.Element]figure.Segment{}
	for _, seg := range figure.Segments() {
		e := s.SegmentElement(seg)
		if !e.Valid() {
			add("%v has invalid element %d", seg, e)
			continue
		}
		if prev, dup := elements[e]; dup {
			add("%v and %v are both %v", prev, seg, e)
		}
		elements[e] = seg
	}

	houses := map[int]position.Position{}
	for _, p := range position.All() {
		h, ok := s.House(p)
		switch {
		case !p.Principal():
			if ok {
				add("%v is not a principal position but is placed in house %d", p, h)
			}
		case !ok:
			add("%v has no house", p)
		case h < 1 || h > 12:
			add("%v is placed in house %d, want 1–12", p, h)
		default:
			if prev, dup := houses[h]; dup {
				add("%v and %v are both placed in house %d", prev, p, h)
			}
			houses[h] = p
		}
	}

	for _, f := range figure.All() {
		if sign, ok := s.Sign(f); ok && !sign.Valid() {
			add("%v has invalid sign %d", f, sign)
		}
		if e, ok := s.Element(f); ok && !e.Valid() {
			add("%v has invalid element %d", f, e)
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("%w %q: %s", ErrInvalid, s.ID(), strings.Join(problems, "; "))
	}
	return nil
}
