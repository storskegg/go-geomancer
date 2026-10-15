package render

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/content"
)

// Header shows the question and when and where the reading took place,
// those that are known, or "" if none are.
//
//	Question  “Will the contract be signed?”
//	When      Thursday, 15 October 2026, 3:30 pm CDT (UTC−05:00)
//	Where     Springfield, IL · 39.8017° N, 89.6437° W · America/Chicago
func Header(question string, c *chart.Circumstances) string {
	var lines []string
	line := func(label, text string) {
		lines = append(lines, lipgloss.NewStyle().Width(Width).Render(labelStyle.Render(fmt.Sprintf("%-9s ", label))+text))
	}
	if question != "" {
		line("Question", nameStyle.Render("“"+question+"”"))
	}
	if c != nil {
		line("When", c.Time.Format("Monday, 2 January 2006, 3:04 pm MST")+mutedStyle.Render(" ("+c.UTCOffset()+")"))
		line("Where", c.Place.Label()+mutedStyle.Render(" · "+c.Place.Coordinates()+" · "+c.TimeZone))
	}
	return strings.Join(lines, "\n")
}

// Meanings interprets the reading: for each house that bears on the
// question (see chart.Reading.Focus), its parts in the question and the
// meaning of the figure in it.
//
//	House 7 · Significator · Partnership
//	  Populus — Many suitors or rivals; …
func Meanings(r chart.Reading, c *content.Set, opts Options) string {
	text := lipgloss.NewStyle().Width(Width).PaddingLeft(2)
	blocks := make([]string, 0, 12)
	for _, f := range r.Focus() {
		parts := []string{labelStyle.Render(fmt.Sprintf("House %d", f.House))}
		if f.Querent {
			parts = append(parts, sigLabelStyle.Render("Querent"))
		}
		if f.Significator {
			parts = append(parts, sigLabelStyle.Render("Significator"))
		}
		if f.Aspect != nil {
			parts = append(parts, opts.aspectLabel(*f.Aspect))
		}
		if sig, ok := c.Signification(f.House); ok {
			parts = append(parts, mutedStyle.Render(sig.Title))
		}

		fig := r.Square[f.House-1].Figure
		meaning := nameStyle.Render(fig.Name()) + " — " + orUnwritten(c.For(fig).House(f.House))
		blocks = append(blocks, strings.Join(parts, mutedStyle.Render(" · "))+"\n"+text.Render(meaning))
	}
	return strings.Join(blocks, "\n\n")
}
