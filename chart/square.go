package chart

import (
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/zodiac"
)

// House is one house of the Astrological Square.
type House struct {
	Number   int               `json:"number"`
	Position position.Position `json:"position"`
	Figure   figure.Figure     `json:"figure"`
	// Sign is the house's sign. It is nil when the system gives no sign to
	// the figure in the 1st House.
	Sign *zodiac.Sign `json:"sign,omitempty"`
	// FigureSign is the sign the system gives the house's figure, or nil if
	// it gives none (as for Caput and Cauda Draconis).
	FigureSign *zodiac.Sign `json:"figureSign,omitempty"`
}

// Square is the Astrological Square. Index i holds House i+1.
type Square [12]House

// BuildSquare places the principal figures of s in houses according to sys.
//
// The 1st House takes the sign sys gives its figure, and each following
// house takes the next sign in zodiacal order. If sys gives the 1st House
// figure no sign, no house has a sign.
func BuildSquare(s Shield, sys system.System) (Square, error) {
	if err := system.Validate(sys); err != nil {
		return Square{}, err
	}

	var sq Square
	for _, p := range position.Principals() {
		h, _ := sys.House(p)
		sq[h-1] = House{Number: h, Position: p, Figure: s.At(p)}
		if sign, ok := sys.Sign(s.At(p)); ok {
			sq[h-1].FigureSign = &sign
		}
	}

	if first, ok := sys.Sign(sq[0].Figure); ok {
		for i := range sq {
			sign := first.Add(i)
			sq[i].Sign = &sign
		}
	}
	return sq, nil
}
