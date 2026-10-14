package render

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"

	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/system"
)

func newTable(headers ...string) *table.Table {
	return table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		Headers(headers...).
		StyleFunc(func(row, _ int) lipgloss.Style {
			s := lipgloss.NewStyle().Padding(0, 1)
			if row == table.HeaderRow {
				return s.Inherit(nameStyle)
			}
			return s
		})
}

// Figures lists the sixteen figures with their correspondences under sys.
func Figures(sys system.System) string {
	t := newTable("Figure", "Pattern", "Planet", "Sign", "Element", "Dots")
	for _, f := range figure.All() {
		sign, hasSign := sys.Sign(f)
		elem, hasElem := sys.Element(f)
		planet, hasPlanet := sys.Planet(f)
		t.Row(
			f.Name(),
			f.Pattern(),
			orNone(planet.String(), hasPlanet),
			orNone(sign.String(), hasSign),
			orNone(elem.String(), hasElem),
			strconv.Itoa(f.Dots()),
		)
	}
	return Heading("Figures · "+sys.Name()) + "\n\n" + t.Render()
}

// Figure describes one figure in full, including its meaning in each house.
func Figure(f figure.Figure, sys system.System, e content.Entry) string {
	sign, hasSign := sys.Sign(f)
	elem, hasElem := sys.Element(f)
	planet, hasPlanet := sys.Planet(f)

	field := func(k, v string) string { return labelStyle.Render(fmt.Sprintf("%-9s", k)) + " " + v }
	facts := strings.Join([]string{
		field("Pattern", f.Pattern()),
		field("Planet", orNone(planet.String(), hasPlanet)),
		field("Sign", orNone(sign.String(), hasSign)),
		field("Element", orNone(elem.String(), hasElem)),
		field("Dots", fmt.Sprintf("%d (%d odd, %d even)", f.Dots(), f.OddDots(), f.EvenDots())),
		field("System", sys.Name()),
	}, "\n")

	header := lipgloss.JoinHorizontal(lipgloss.Top, Glyph(f), "    ", facts)

	houses := newTable("House", "Meaning").Width(Width)
	for n := 1; n <= 12; n++ {
		houses.Row(strconv.Itoa(n), orUnwritten(e.House(n)))
	}

	return strings.Join([]string{
		Heading(f.Name()),
		header,
		labelStyle.Render("In brief") + "\n" + text(e.Short),
		labelStyle.Render("Description") + "\n" + text(e.Long),
		houses.Render(),
	}, "\n\n")
}

// Systems lists the registered systems and their segment elements.
func Systems(systems []system.System) string {
	t := newTable("ID", "Name", "Head", "Neck", "Body", "Feet")
	for _, s := range systems {
		row := []string{s.ID(), s.Name()}
		for _, seg := range figure.Segments() {
			row = append(row, s.SegmentElement(seg).String())
		}
		t.Row(row...)
	}
	return Heading("Systems") + "\n\n" + t.Render()
}

// orUnwritten returns s, or a note that it hasn't been written yet.
func orUnwritten(s string) string {
	if s == "" {
		return mutedStyle.Render("Not yet written.")
	}
	return s
}

// text wraps s to Width, or notes that it hasn't been written.
func text(s string) string {
	return lipgloss.NewStyle().Width(Width).Render(orUnwritten(s))
}
