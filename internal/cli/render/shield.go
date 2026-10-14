package render

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/position"
)

// cellWidth is the width of a Mother or Daughter cell, wide enough for the
// longest single-word name. Each row below is twice as wide per figure, so
// every figure sits centred beneath the two it was summed from.
const cellWidth = 11

// span returns the width of a cell covering n Mother/Daughter columns.
func span(n int) int { return n * cellWidth }

// Shield draws the Shield in its traditional layout: Mothers and Daughters
// across the top reading right to left (M1 at the far right), then the
// Nieces, the Witnesses, and the Judge with the Judgement to its right.
func Shield(s chart.Shield, opts Options) string {
	row := func(width int, ps ...position.Position) string {
		cells := make([]string, 0, len(ps))
		for _, p := range ps {
			cells = append(cells, shieldCell(s, p, width))
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, cells...)
	}

	judge := lipgloss.JoinHorizontal(lipgloss.Top,
		strings.Repeat(" ", span(2)),
		shieldCell(s, position.J, span(4)),
		shieldCell(s, position.JM, span(2)),
	)

	legend := mutedStyle.Render(strings.Join([]string{
		"M Mothers", "D Daughters", "N " + opts.nieces(), "W Witnesses", "J Judge", "JM Judgement",
	}, " · "))

	return strings.Join([]string{
		row(span(1),
			position.D4, position.D3, position.D2, position.D1,
			position.M4, position.M3, position.M2, position.M1),
		row(span(2), position.N4, position.N3, position.N2, position.N1),
		row(span(4), position.WL, position.WR),
		judge,
		legend,
	}, "\n\n")
}

func shieldCell(s chart.Shield, p position.Position, width int) string {
	f := s.At(p)
	// Joined with plain newlines rather than JoinVertical, which pads lines to
	// a common width and would pull the dots off-centre beside a long name.
	// Align centres each line within the cell on its own.
	body := strings.Join([]string{labelStyle.Render(p.String()), Glyph(f), nameStyle.Render(f.Name())}, "\n")
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(body)
}
