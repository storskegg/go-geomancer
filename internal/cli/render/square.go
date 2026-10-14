package render

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/storskegg/go-geomancer/chart"
)

// houseWidth includes the border; four houses match the Shield's width.
const houseWidth = 2 * cellWidth

// Square draws the Astrological Square as a grid of houses, four per row.
func Square(sq chart.Square) string {
	cell := lipgloss.NewStyle().
		Width(houseWidth).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderStyle.GetForeground())

	rows := make([]string, 0, 3)
	for r := range 3 {
		cells := make([]string, 0, 4)
		for _, h := range sq[r*4 : r*4+4] {
			sign := mutedStyle.Render(none)
			if h.Sign != nil {
				sign = h.Sign.String()
			}
			cells = append(cells, cell.Render(strings.Join([]string{
				labelStyle.Render(fmt.Sprintf("House %d", h.Number)),
				sign,
				"",
				Glyph(h.Figure),
				"",
				nameStyle.Render(h.Figure.Name()),
				mutedStyle.Render(h.Position.String()),
			}, "\n")))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cells...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
