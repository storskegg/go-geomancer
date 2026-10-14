package render

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/zodiac"
)

// houseWidth includes the border.
const houseWidth = 2 * cellWidth

// SquareWidth is the width of the Square: five houses across.
const SquareWidth = 5 * houseWidth

// squareCells gives each house's row and column in the Square's 5×5 grid;
// index i is House i+1. The houses run counterclockwise around the edge,
// three to a side, from House 1 at the middle of the left side, as in an
// astrological chart with the Ascendant on the left. The corners and the
// 3×3 centre are empty:
//
//	    11  10   9
//	12              8
//	 1              7
//	 2              6
//	     3   4   5
var squareCells = [12][2]int{
	{2, 0}, {3, 0}, // left side, middle to bottom
	{4, 1}, {4, 2}, {4, 3}, // bottom, left to right
	{3, 4}, {2, 4}, {1, 4}, // right side, bottom to top
	{0, 3}, {0, 2}, {0, 1}, // top, right to left
	{1, 0}, // left side, top
}

// A house's inner width splits into a column at each edge for symbols and
// a centre for text, with a space either side of the centre. The centre
// fits the longest figure name, "Caput Draconis".
const (
	innerWidth  = houseWidth - 2 // less the border
	edgeWidth   = 2              // a sign symbol is two columns wide
	centreWidth = innerWidth - 2*edgeWidth - 2
)

// Square draws the Astrological Square, its houses arranged around the edge
// of a 5×5 grid (see squareCells), with a legend for the symbols.
//
// Each house shows its number and sign, the figure, and the figure's name
// and position. The house's sign symbol and element flank the sign's name;
// the figure's sign symbol and element flank the figure's name. The right
// edge of the position row is free for a further element.
func Square(sq chart.Square, opts Options) string {
	cell := lipgloss.NewStyle().
		Width(houseWidth).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderStyle.GetForeground())

	var grid [5][5]string
	for i, h := range sq {
		signName := mutedStyle.Render(none)
		if h.Sign != nil {
			signName = h.Sign.String()
		}
		r, c := squareCells[i][0], squareCells[i][1]
		grid[r][c] = cell.Render(strings.Join([]string{
			labelStyle.Render(fmt.Sprintf("House %d", h.Number)),
			opts.signRow(h.Sign, signName),
			"",
			Glyph(h.Figure),
			"",
			opts.signRow(h.FigureSign, nameStyle.Render(h.Figure.Name())),
			edgeRow("", mutedStyle.Render(h.Position.String()), ""),
		}, "\n"))
	}

	// Empty cells are blank blocks the size of a house.
	blankLine := strings.Repeat(" ", houseWidth)
	blank := strings.TrimSuffix(strings.Repeat(blankLine+"\n", lipgloss.Height(grid[2][0])), "\n")

	rows := make([]string, 0, len(grid))
	for _, row := range grid {
		cells := make([]string, 0, len(row))
		for _, c := range row {
			if c == "" {
				c = blank
			}
			cells = append(cells, c)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...) + "\n" + opts.squareLegend()
}

// signRow puts sign's symbol at the left edge and its element at the right
// edge of centre. Both edges are blank if sign is nil.
func (o Options) signRow(sign *zodiac.Sign, centre string) string {
	if sign == nil {
		return edgeRow("", centre, "")
	}
	return edgeRow(o.signSymbol(*sign), centre, o.elementSymbol(sign.Element()))
}

// edgeRow lays out one row of a house: left flush left, centre centred,
// right flush right.
func edgeRow(left, centre, right string) string {
	return lipgloss.PlaceHorizontal(edgeWidth, lipgloss.Left, left) + " " +
		lipgloss.PlaceHorizontal(centreWidth, lipgloss.Center, centre) + " " +
		lipgloss.PlaceHorizontal(edgeWidth, lipgloss.Right, right)
}

func (o Options) squareLegend() string {
	items := make([]string, 0, 4)
	for _, e := range zodiac.Elements() {
		items = append(items, o.elementSymbol(e)+" "+e.String())
	}
	return mutedStyle.Render("Left: sign · Right: its element · ") + strings.Join(items, mutedStyle.Render(" · "))
}
