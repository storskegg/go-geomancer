// Package render draws charts and figures for the terminal with Lip Gloss.
//
// Functions return styled strings; the caller writes them through a
// colorprofile writer, which strips or downsamples styling to suit the
// output.
package render

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/storskegg/go-geomancer/figure"
)

// Basic ANSI colours, so output follows the terminal's own palette.
var (
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	nameStyle    = lipgloss.NewStyle().Bold(true)
	dotStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	mutedStyle   = lipgloss.NewStyle().Faint(true)
	headingStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	borderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// Width is the width of the widest output: the Shield and the Square.
const Width = 8 * cellWidth

const (
	oddRow  = "  ●  "
	evenRow = "●   ●"
	none    = "—"
)

// Options adjust labels without changing the chart.
type Options struct {
	// Nephews calls N1–N4 Nephews rather than Nieces.
	Nephews bool
}

func (o Options) nieces() string {
	if o.Nephews {
		return "Nephews"
	}
	return "Nieces"
}

// Glyph draws a figure as four rows of dots, top to bottom.
func Glyph(f figure.Figure) string {
	rows := make([]string, 0, 4)
	for _, p := range f.Segments() {
		row := evenRow
		if p == figure.Odd {
			row = oddRow
		}
		rows = append(rows, dotStyle.Render(row))
	}
	return strings.Join(rows, "\n")
}

// Heading renders a section heading.
func Heading(s string) string { return headingStyle.Render(s) }

func orNone(s string, ok bool) string {
	if !ok {
		return mutedStyle.Render(none)
	}
	return s
}
