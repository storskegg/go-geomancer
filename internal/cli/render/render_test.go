package render

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/system/traditional"
)

func TestGlyph(t *testing.T) {
	got := ansi.Strip(Glyph(figure.Amissio)) // 1212
	want := strings.Join([]string{oddRow, evenRow, oddRow, evenRow}, "\n")
	if got != want {
		t.Errorf("Glyph(Amissio) =\n%s\nwant\n%s", got, want)
	}
}

func TestWidths(t *testing.T) {
	s, err := chart.Cast([4]figure.Figure{figure.CaputDraconis, figure.Conjunctio, figure.Acquisitio, figure.FortunaMajor})
	if err != nil {
		t.Fatal(err)
	}
	sq, err := chart.BuildSquare(s, traditional.System)
	if err != nil {
		t.Fatal(err)
	}
	e := content.Entry{Figure: figure.Puer, Long: strings.Repeat("A long description that must wrap. ", 20)}
	e.Houses[0] = strings.Repeat("A long house meaning that must wrap. ", 10)

	for name, out := range map[string]string{
		"shield": Shield(s, Options{}),
		"square": Square(sq),
		"figure": Figure(figure.Puer, traditional.System, e),
	} {
		if w := lipgloss.Width(out); w > Width {
			t.Errorf("%s is %d columns wide, want at most %d", name, w, Width)
		}
	}
}

func TestNephews(t *testing.T) {
	s, _ := chart.Cast([4]figure.Figure{})
	if out := Shield(s, Options{Nephews: true}); !strings.Contains(out, "N Nephews") || strings.Contains(out, "Nieces") {
		t.Error("Nephews option not applied to the legend")
	}
}
