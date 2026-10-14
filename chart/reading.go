package chart

import (
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/system"
)

// Reading is a complete chart: the Shield and the Astrological Square cast
// under a System.
type Reading struct {
	System string `json:"system"`
	// Significator is the house (1–12) of the matter asked about, or 0 if
	// the reading has none. House 1 always signifies the querent.
	Significator int    `json:"significator,omitempty"`
	Shield       Shield `json:"shield"`
	Square       Square `json:"square"`
}

// NewReading casts the Shield from the Mothers and builds the Astrological
// Square under sys.
func NewReading(mothers [4]figure.Figure, sys system.System) (Reading, error) {
	shield, err := Cast(mothers)
	if err != nil {
		return Reading{}, err
	}
	square, err := BuildSquare(shield, sys)
	if err != nil {
		return Reading{}, err
	}
	return Reading{System: sys.ID(), Shield: shield, Square: square}, nil
}
