package render

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/storskegg/go-geomancer/chart"
)

var mixedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow

// Timing lists the judgements of the moment against the matter, each marked
// by whether it helps (✓), hinders (✗), is mixed (±) or neither (·).
func Timing(js []chart.Judgement, opts Options) string {
	text := lipgloss.NewStyle().Width(Width - 4)
	lines := make([]string, len(js))
	for i, j := range js {
		mark := opts.effectMark(j.Effect)
		lines[i] = lipgloss.JoinHorizontal(lipgloss.Top, mark+" ", labelStyle.Render(padRight(j.About, 5))+" ", text.Render(j.Text))
	}
	return strings.Join(lines, "\n")
}

func (o Options) effectMark(e chart.Effect) string {
	marks, ascii := []string{"✓", "✗", "±", "·"}, []string{"+", "-", "~", "."}
	if int(e) >= len(marks) {
		return " "
	}
	m := marks[e]
	if o.Symbols == ASCII {
		m = ascii[e]
	}
	switch e {
	case chart.Helps:
		return easingStyle.Render(m)
	case chart.Hinders:
		return stressfulStyle.Render(m)
	case chart.Mixed:
		return mixedStyle.Render(m)
	case chart.Neutral:
	}
	return mutedStyle.Render(m)
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}
