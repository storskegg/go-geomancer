// Package chart casts geomantic charts: the Shield derived from four Mothers
// and, given a System, the Astrological Square.
package chart

import (
	"errors"
	"fmt"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/position"
)

// ErrInvalidMother is returned by Cast when a Mother isn't a valid figure.
var ErrInvalidMother = errors.New("chart: invalid mother")

// Shield holds every figure derived from the four Mothers.
type Shield struct {
	Mothers      [4]figure.Figure `json:"mothers"`
	Daughters    [4]figure.Figure `json:"daughters"`
	Nieces       [4]figure.Figure `json:"nieces"`
	RightWitness figure.Figure    `json:"rightWitness"`
	LeftWitness  figure.Figure    `json:"leftWitness"`
	Judge        figure.Figure    `json:"judge"`
	Judgement    figure.Figure    `json:"judgement"`
}

// Cast derives a Shield from the four Mothers:
//
//   - Daughter i takes segment i of each Mother, M1 through M4, as its segments
//     top to bottom.
//   - N1 = M1+M2, N2 = M3+M4, N3 = D1+D2, N4 = D3+D4
//   - WR = N1+N2, WL = N3+N4
//   - J = WR+WL, JM = J+M1
func Cast(mothers [4]figure.Figure) (Shield, error) {
	for i, m := range mothers {
		if !m.Valid() {
			return Shield{}, fmt.Errorf("%w: M%d = %d", ErrInvalidMother, i+1, uint8(m))
		}
	}

	s := Shield{Mothers: mothers}
	for _, seg := range figure.Segments() {
		var p [4]figure.Parity
		for i, m := range mothers {
			p[i] = m.Segment(seg)
		}
		s.Daughters[seg] = figure.FromSegments(p)
	}

	m, d := s.Mothers, s.Daughters
	s.Nieces = [4]figure.Figure{m[0].Add(m[1]), m[2].Add(m[3]), d[0].Add(d[1]), d[2].Add(d[3])}

	n := s.Nieces
	s.RightWitness = n[0].Add(n[1])
	s.LeftWitness = n[2].Add(n[3])
	s.Judge = s.RightWitness.Add(s.LeftWitness)
	s.Judgement = s.Judge.Add(m[0])
	return s, nil
}

// At returns the figure at position p. It panics if p is not a valid
// position.
func (s Shield) At(p position.Position) figure.Figure {
	switch p {
	case position.M1, position.M2, position.M3, position.M4:
		return s.Mothers[p-position.M1]
	case position.D1, position.D2, position.D3, position.D4:
		return s.Daughters[p-position.D1]
	case position.N1, position.N2, position.N3, position.N4:
		return s.Nieces[p-position.N1]
	case position.WR:
		return s.RightWitness
	case position.WL:
		return s.LeftWitness
	case position.J:
		return s.Judge
	case position.JM:
		return s.Judgement
	}
	panic(fmt.Sprintf("chart: invalid position %v", p))
}
