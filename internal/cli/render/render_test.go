package render

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/system/traditional"
	"github.com/storskegg/go-geomancer/zodiac"
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

func plain(s string) string { return ansi.Strip(s) }

func TestFigures(t *testing.T) {
	out := plain(Figures(traditional.System))
	if !strings.Contains(out, "Figures · Traditional") {
		t.Errorf("missing heading:\n%s", out)
	}
	for _, f := range figure.All() {
		if !strings.Contains(out, f.Name()) || !strings.Contains(out, f.Pattern()) {
			t.Errorf("table lacks %v (%s)", f, f.Pattern())
		}
	}
	// Caput Draconis has no sign: its row shows a dash in the Sign column.
	for line := range strings.SplitSeq(out, "\n") {
		if strings.Contains(line, "Caput Draconis") && !strings.Contains(line, "│ —") {
			t.Errorf("Caput Draconis row lacks a dash: %q", line)
		}
	}
}

func TestSystems(t *testing.T) {
	out := plain(Systems([]system.System{traditional.System}))
	for _, want := range []string{"Systems", "traditional", "Traditional", "Fire", "Air", "Water", "Earth"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestSquareSigns(t *testing.T) {
	s, _ := chart.Cast([4]figure.Figure{figure.Amissio, figure.FortunaMinor, figure.Via, figure.Populus})
	sq, err := chart.BuildSquare(s, traditional.System)
	if err != nil {
		t.Fatal(err)
	}
	out := plain(Square(sq))
	for _, h := range sq {
		if !strings.Contains(out, fmt.Sprintf("House %d ", h.Number)) || !strings.Contains(out, h.Sign.String()) {
			t.Errorf("square lacks house %d or its sign %v", h.Number, h.Sign)
		}
	}

	// A square with no signs shows dashes.
	for i := range sq {
		sq[i].Sign = nil
	}
	if out := plain(Square(sq)); strings.Count(out, "—") != 12 {
		t.Errorf("unsigned square has %d dashes, want 12", strings.Count(out, "—"))
	}
}

func TestFigureDetail(t *testing.T) {
	e := content.Entry{Figure: figure.Puer, Short: "Rash action."}
	e.Houses[6] = "Quarrels in partnership."
	out := plain(Figure(figure.Puer, traditional.System, e))
	for _, want := range []string{
		"Puer", "1121", "Mars", "Aries", "Fire", "5 (3 odd, 2 even)", "Traditional",
		"Rash action.", "Quarrels in partnership.", "Not yet written.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("detail lacks %q:\n%s", want, out)
		}
	}
	// One unwritten description plus eleven unwritten houses.
	if got := strings.Count(out, "Not yet written."); got != 12 {
		t.Errorf("%d placeholders, want 12", got)
	}
}

func TestShieldLayout(t *testing.T) {
	s, _ := chart.Cast([4]figure.Figure{figure.Amissio, figure.FortunaMinor, figure.Via, figure.Populus})
	lines := strings.Split(plain(Shield(s, Options{})), "\n")
	// The top row reads right to left: D4 first, M1 last.
	top := strings.Fields(lines[0])
	want := []string{"D4", "D3", "D2", "D1", "M4", "M3", "M2", "M1"}
	if strings.Join(top, " ") != strings.Join(want, " ") {
		t.Errorf("top row = %v, want %v", top, want)
	}
	// Each label sits centred over its column.
	for i, label := range want {
		col := strings.Index(lines[0], label)
		if centre := i*cellWidth + cellWidth/2; col < centre-1 || col > centre {
			t.Errorf("%s at column %d, want about %d", label, col, centre)
		}
	}
	if out := plain(Shield(s, Options{})); !strings.Contains(out, "N Nieces") {
		t.Error("default legend lacks Nieces")
	}
}

func TestOrNone(t *testing.T) {
	if got := orNone("Leo", true); got != "Leo" {
		t.Errorf("orNone(Leo, true) = %q", got)
	}
	if got := plain(orNone("Leo", false)); got != none {
		t.Errorf("orNone(Leo, false) = %q", got)
	}
}

func TestHeading(t *testing.T) {
	if got := plain(Heading("The Shield")); got != "The Shield" {
		t.Errorf("Heading = %q", got)
	}
}

// A system that gives no planets shows dashes rather than a planet.
func TestUnmappedPlanet(t *testing.T) {
	bare := &system.Mapping{
		Slug:            "bare",
		Title:           "Bare",
		SegmentElements: [4]zodiac.Element{zodiac.Fire, zodiac.Air, zodiac.Water, zodiac.Earth},
		Houses:          [12]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
	}
	for line := range strings.SplitSeq(plain(Figures(bare)), "\n") {
		if strings.Contains(line, "Fortuna Major") && strings.Count(line, "—") != 3 {
			t.Errorf("row %q: want dashes for planet, sign and element", line)
		}
	}
	if out := plain(Figure(figure.Puer, bare, content.Entry{Figure: figure.Puer})); !strings.Contains(out, "Planet    —") {
		t.Errorf("detail lacks an unmapped planet:\n%s", out)
	}
	if out := plain(Figure(figure.CaputDraconis, traditional.System, content.Entry{})); !strings.Contains(out, "Planet    North Node") || !strings.Contains(out, "Sign      —") {
		t.Errorf("Caput Draconis detail:\n%s", out)
	}
}
