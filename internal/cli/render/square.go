package render

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/zodiac"
)

// houseWidth includes the border, and houseGap separates neighbouring
// houses.
const (
	houseWidth = 18
	houseGap   = 1
)

// SquareWidth is the width of the Square: five houses across.
const SquareWidth = 5*houseWidth + 4*houseGap

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

// Rows that carry symbols split a house's inner width into a column at
// each edge and a centre. The left column fits a sign symbol, which is two
// columns wide; the right fits an element symbol, which is one. The centre
// fits the longest sign name, "Sagittarius", with a space either side.
const (
	innerWidth  = houseWidth - 2 // less the border
	leftWidth   = 2
	rightWidth  = 1
	centreWidth = innerWidth - leftWidth - rightWidth
)

// Square draws the Astrological Square, its houses arranged around the edge
// of a 5×5 grid (see squareCells), with a legend for the symbols. The
// significator's house (1–12; 0 for none) has a heavy border.
//
// Each house shows its number, its sign, the figure, and the figure's name
// and position. Signs' symbols run down the left edge and elements down the
// right: the house's sign symbol and element flank the sign's name, and the
// figure's sign symbol and element flank the top row of its dots. The right
// edge of the second row of dots is free for a further element.
func Square(sq chart.Square, significator int, opts Options) string {
	cell := lipgloss.NewStyle().
		Width(houseWidth).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderStyle.GetForeground())
	sigCell := cell.Border(lipgloss.ThickBorder()).BorderForeground(headingStyle.GetForeground())

	var grid [5][5]string
	for i, h := range sq {
		signName := mutedStyle.Render(none)
		if h.Sign != nil {
			signName = h.Sign.String()
		}
		// Every row of dots goes through edgeRow, so they stay aligned
		// whether or not they carry symbols.
		dots := strings.Split(Glyph(h.Figure), "\n")
		dots[0] = opts.signRow(h.FigureSign, dots[0])
		for j := 1; j < len(dots); j++ {
			dots[j] = edgeRow("", dots[j], "")
		}
		lines := append([]string{
			labelStyle.Render(fmt.Sprintf("House %d", h.Number)),
			opts.signRow(h.Sign, signName),
			"",
		}, dots...)
		lines = append(lines,
			"",
			nameStyle.Render(h.Figure.Name()),
			mutedStyle.Render(h.Position.String()),
		)
		style := cell
		if h.Number == significator {
			style = sigCell
		}
		r, c := squareCells[i][0], squareCells[i][1]
		grid[r][c] = style.Render(strings.Join(lines, "\n"))
	}

	// Empty cells are blank blocks the size of a house.
	blankLine := strings.Repeat(" ", houseWidth)
	blank := strings.TrimSuffix(strings.Repeat(blankLine+"\n", lipgloss.Height(grid[2][0])), "\n")
	gap := strings.Repeat(" ", houseGap)

	rows := make([]string, 0, len(grid))
	for _, row := range grid {
		cells := make([]string, 0, 2*len(row))
		for j, c := range row {
			if j > 0 {
				cells = append(cells, gap)
			}
			if c == "" {
				c = blank
			}
			cells = append(cells, c)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...) + "\n" + opts.squareLegend(significator != 0)
}

// Significator describes the significator, for display above the Square.
func Significator(sig content.Signification) string {
	return labelStyle.Render(fmt.Sprintf("Significator: House %d · %s", sig.House, sig.Title)) +
		mutedStyle.Render(" — "+sig.Text)
}

// signRow puts sign's symbol at the left edge and its element at the right
// edge of centre. Both edges are blank if sign is nil.
func (o Options) signRow(sign *zodiac.Sign, centre string) string {
	if sign == nil {
		return edgeRow("", centre, "")
	}
	return edgeRow(o.signSymbol(*sign), centre, o.elementSymbol(sign.Element()))
}

// edgeRow lays out one row of a house across its inner width: left flush
// left, centre centred, right flush right.
func edgeRow(left, centre, right string) string {
	return lipgloss.PlaceHorizontal(leftWidth, lipgloss.Left, left) +
		lipgloss.PlaceHorizontal(centreWidth, lipgloss.Center, centre) +
		lipgloss.PlaceHorizontal(rightWidth, lipgloss.Right, right)
}

func (o Options) squareLegend(significator bool) string {
	items := make([]string, 0, 4)
	for _, e := range zodiac.Elements() {
		items = append(items, o.elementSymbol(e)+" "+e.String())
	}
	legend := mutedStyle.Render("Left: sign · Right: its element · ") + strings.Join(items, mutedStyle.Render(" · "))
	if significator {
		legend += "\n" + mutedStyle.Render("Heavy border: significator")
	}
	return legend
}
