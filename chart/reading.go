package chart

import (
	"github.com/storskegg/go-geomancer/aspect"
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/system"
)

// Reading is a complete chart: the Shield and the Astrological Square cast
// under a System.
type Reading struct {
	System string `json:"system"`
	// Question is the querent's question in their own words, if given.
	Question string `json:"question,omitempty"`
	// Significator is the house (1–12) of the matter asked about, or 0 if
	// the reading has none. House 1 always signifies the querent. Set it
	// with SetSignificator.
	Significator int `json:"significator,omitempty"`
	// Aspects are the houses aspecting the significator's; see
	// aspect.Houses.
	Aspects []aspect.House `json:"aspects,omitempty"`
	Shield  Shield         `json:"shield"`
	Square  Square         `json:"square"`
}

// SetSignificator sets the significator to house n (1–12) and works out the
// houses aspecting it. n = 0 clears both.
func (r *Reading) SetSignificator(n int) error {
	if n == 0 {
		r.Significator, r.Aspects = 0, nil
		return nil
	}
	aspects, err := aspect.Houses(n)
	if err != nil {
		return err
	}
	r.Significator, r.Aspects = n, aspects
	return nil
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
