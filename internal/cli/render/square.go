package render

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/zodiac"
)

// houseWidth includes the border; four houses match the Shield's width.
const houseWidth = 2 * cellWidth

// A house's inner width splits into a column at each edge for symbols and
// a centre for text, with a space either side of the centre. The centre
// fits the longest figure name, "Caput Draconis".
const (
	innerWidth  = houseWidth - 2 // less the border
	edgeWidth   = 2              // a sign symbol is two columns wide
	centreWidth = innerWidth - 2*edgeWidth - 2
)

// Square draws the Astrological Square as a grid of houses, four per row,
// with a legend for the symbols.
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

	rows := make([]string, 0, 3)
	for r := range 3 {
		cells := make([]string, 0, 4)
		for _, h := range sq[r*4 : r*4+4] {
			signName := mutedStyle.Render(none)
			if h.Sign != nil {
				signName = h.Sign.String()
			}
			cells = append(cells, cell.Render(strings.Join([]string{
				labelStyle.Render(fmt.Sprintf("House %d", h.Number)),
				opts.signRow(h.Sign, signName),
				"",
				Glyph(h.Figure),
				"",
				opts.signRow(h.FigureSign, nameStyle.Render(h.Figure.Name())),
				edgeRow("", mutedStyle.Render(h.Position.String()), ""),
			}, "\n")))
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
