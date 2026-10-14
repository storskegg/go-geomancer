package render

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/storskegg/go-geomancer/aspect"
)

var (
	stressfulStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
	easingStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	sigLabelStyle  = headingStyle
)

// aspectStyle colours an aspect by its nature: stressful or easing.
func aspectStyle(a aspect.Aspect) lipgloss.Style {
	if a.Stressful() {
		return stressfulStyle
	}
	return easingStyle
}

// sideMark marks an aspect's side: ↻ Dexter (clockwise), ↺ Sinister, or in
// ASCII "dex" and "sin". The Opposition has no side.
func (o Options) sideMark(s aspect.Side) string {
	switch s {
	case aspect.Dexter:
		if o.Symbols == ASCII {
			return "dex"
		}
		return "↻"
	case aspect.Sinister:
		if o.Symbols == ASCII {
			return "sin"
		}
		return "↺"
	case aspect.Neither:
	}
	return ""
}

func (o Options) aspectSymbol(a aspect.Aspect) string {
	if o.Symbols == ASCII {
		return ""
	}
	return a.Symbol()
}

// aspectLabel names a house's aspect to the significator, such as
// "□ Square ↻", coloured by its nature.
func (o Options) aspectLabel(h aspect.House) string {
	parts := []string{o.aspectSymbol(h.Aspect), h.Aspect.String(), o.sideMark(h.Side)}
	return aspectStyle(h.Aspect).Render(strings.Join(nonEmpty(parts), " "))
}

// Aspects lists the houses aspecting the significator, for display above
// the Square: "Opposition 4 · Square 7 ↻ 1 ↺ · …".
func Aspects(houses []aspect.House, opts Options) string {
	var groups []string
	for _, a := range aspect.All() {
		var hs []string
		for _, h := range houses {
			if h.Aspect == a {
				hs = append(hs, strings.Join(nonEmpty([]string{fmt.Sprint(h.House), opts.sideMark(h.Side)}), " "))
			}
		}
		if len(hs) > 0 {
			name := strings.Join(nonEmpty([]string{opts.aspectSymbol(a), a.String()}), " ")
			groups = append(groups, aspectStyle(a).Render(name)+" "+strings.Join(hs, ", "))
		}
	}
	return labelStyle.Render("Aspects: ") + strings.Join(groups, mutedStyle.Render(" · "))
}

// aspectLegend explains the significator's border and the aspect labels.
func (o Options) aspectLegend() string {
	name := func(a aspect.Aspect) string {
		return strings.Join(nonEmpty([]string{o.aspectSymbol(a), a.String()}), " ")
	}
	nature := mutedStyle.Render("Heavy border: significator · ") +
		stressfulStyle.Render(name(aspect.Opposition)+", "+name(aspect.Square)+": stressful") +
		mutedStyle.Render(" · ") +
		easingStyle.Render(name(aspect.Trine)+", "+name(aspect.Sextile)+": easing")
	sides := mutedStyle.Render(o.sideMark(aspect.Dexter) + " Dexter (clockwise), the stronger of a pair · " +
		o.sideMark(aspect.Sinister) + " Sinister")
	return nature + "\n" + sides
}

func nonEmpty(ss []string) []string {
	out := ss[:0:0]
	for _, s := range ss {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
