// Package render draws charts and figures for the terminal with Lip Gloss.
//
// Functions return styled strings; the caller writes them through a
// colorprofile writer, which strips or downsamples styling to suit the
// output.
package render

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/zodiac"
)

// Basic ANSI colours, so output follows the terminal's own palette.
var (
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	nameStyle    = lipgloss.NewStyle().Bold(true)
	dotStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	mutedStyle   = lipgloss.NewStyle().Faint(true)
	headingStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	borderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	// elementStyles is indexed by zodiac.Element: Fire, Air, Water, Earth.
	elementStyles = [4]lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color("1")), // red
		lipgloss.NewStyle().Foreground(lipgloss.Color("3")), // yellow
		lipgloss.NewStyle().Foreground(lipgloss.Color("4")), // blue
		lipgloss.NewStyle().Foreground(lipgloss.Color("2")), // green
	}
)

// Width is the width of the widest output: the Shield and the Square.
const Width = 8 * cellWidth

const (
	oddRow  = "  ●  "
	evenRow = "●   ●"
	none    = "—"
)

// Options adjust labels and symbols without changing the chart.
type Options struct {
	// Nephews calls N1–N4 Nephews rather than Nieces.
	Nephews bool
	// Symbols chooses how signs and elements are drawn.
	Symbols Symbols
}

// Symbols is a set of symbols for signs and elements.
type Symbols uint8

const (
	// Unicode draws signs as ♈–♓ and elements as alchemical triangles
	// (🜂 🜁 🜄 🜃). Many terminal fonts lack the triangles.
	Unicode Symbols = iota
	// ASCII draws signs as two-letter abbreviations (Ar, Ta, …) and
	// elements as letters (F, A, W, E).
	ASCII
)

var symbolNames = []string{"unicode", "ascii"}

// ParseSymbols parses "unicode" or "ascii", ignoring case.
func ParseSymbols(s string) (Symbols, error) {
	for i, n := range symbolNames {
		if strings.EqualFold(s, n) {
			return Symbols(i), nil
		}
	}
	return 0, fmt.Errorf("unknown symbols %q: want unicode or ascii", s)
}

func (s Symbols) String() string {
	if int(s) < len(symbolNames) {
		return symbolNames[s]
	}
	return fmt.Sprintf("Symbols(%d)", s)
}

func (o Options) signSymbol(s zodiac.Sign) string {
	if o.Symbols == ASCII {
		return s.Abbreviation()
	}
	return s.Symbol()
}

// elementSymbol draws e in its colour.
func (o Options) elementSymbol(e zodiac.Element) string {
	sym := e.Symbol()
	if o.Symbols == ASCII {
		sym = e.String()[:1]
	}
	return elementStyles[e].Render(sym)
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
