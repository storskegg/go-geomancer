// Package input supplies the four Mothers for a reading.
package input

import (
	"fmt"

	"github.com/storskegg/go-geomancer/figure"
)

// MothersSource supplies the four Mothers, M1 through M4.
type MothersSource interface {
	Mothers() ([4]figure.Figure, error)
}

// Args reads the Mothers from four arguments, each in any form accepted by
// figure.Parse.
type Args []string

func (a Args) Mothers() ([4]figure.Figure, error) {
	var m [4]figure.Figure
	if len(a) != len(m) {
		return m, fmt.Errorf("expected 4 mothers, got %d", len(a))
	}
	for i, s := range a {
		f, err := figure.Parse(s)
		if err != nil {
			return m, fmt.Errorf("mother M%d: %w", i+1, err)
		}
		m[i] = f
	}
	return m, nil
}
