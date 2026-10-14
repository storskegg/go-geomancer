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
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/system/goldendawn"
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

	for name, tt := range map[string]struct {
		out string
		max int
	}{
		"shield": {Shield(s, Options{}), Width},
		"square": {Square(sq, Options{}), SquareWidth},
		"ascii":  {Square(sq, Options{Symbols: ASCII}), SquareWidth},
		"figure": {Figure(figure.Puer, traditional.System, e), Width},
	} {
		if w := lipgloss.Width(tt.out); w > tt.max {
			t.Errorf("%s is %d columns wide, want at most %d", name, w, tt.max)
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
	// Golden Dawn gives Caput Draconis no sign: its row shows a dash.
	for line := range strings.SplitSeq(plain(Figures(goldendawn.System)), "\n") {
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
	out := plain(Square(sq, Options{}))
	for _, h := range sq {
		if !strings.Contains(out, fmt.Sprintf("House %d ", h.Number)) || !strings.Contains(out, h.Sign.String()) {
			t.Errorf("square lacks house %d or its sign %v", h.Number, h.Sign)
		}
	}

	// A square with no signs shows dashes.
	for i := range sq {
		sq[i].Sign = nil
	}
	if out := plain(Square(sq, Options{})); strings.Count(out, "—") != 12 {
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
	if out := plain(Figure(figure.CaputDraconis, goldendawn.System, content.Entry{})); !strings.Contains(out, "Planet    North Node") || !strings.Contains(out, "Sign      —") {
		t.Errorf("Caput Draconis detail:\n%s", out)
	}
}

// tile returns the plain lines of house n's tile (1–12) in a rendered Square.
func tile(out string, n int) []string {
	lines := strings.Split(plain(out), "\n")
	row, col := squareCells[n-1][0], squareCells[n-1][1]
	const height = 12 // border, 10 lines, border
	var t []string
	for _, l := range lines[row*height : row*height+height] {
		t = append(t, ansi.Cut(l, col*houseWidth, (col+1)*houseWidth))
	}
	return t
}

func testSquare() chart.Square {
	taurus, libra, leo := zodiac.Taurus, zodiac.Libra, zodiac.Leo
	var sq chart.Square
	for i := range sq {
		sq[i] = chart.House{Number: i + 1, Position: position.Principals()[i], Figure: figure.Populus, Sign: &leo}
	}
	sq[0] = chart.House{Number: 1, Position: position.M1, Figure: figure.Amissio, Sign: &taurus, FigureSign: &libra}
	sq[1] = chart.House{Number: 2, Position: position.M2, Figure: figure.CaudaDraconis, Sign: &leo}
	sq[2] = chart.House{Number: 3, Position: position.M3, Figure: figure.Via}
	return sq
}

func TestSquareTile(t *testing.T) {
	out := Square(testSquare(), Options{})
	tests := []struct {
		house int
		row   int // line within the tile; 0 is the top border
		want  string
	}{
		// The house's sign symbol and element flank the sign's name.
		{1, 2, "│♉     Taurus      🜃│"},
		// The figure's sign symbol and element flank the figure's name.
		{1, 9, "│♎    Amissio      🜁│"},
		// The position row leaves the right edge free.
		{1, 10, "│         M1         │"},
		// A figure without a sign has blank edges.
		{2, 2, "│♌      Leo        🜂│"},
		{2, 9, "│   Cauda Draconis   │"},
		// A house without a sign shows a dash and blank edges.
		{3, 2, "│         —          │"},
	}
	for _, tt := range tests {
		lines := tile(out, tt.house)
		if got := lines[tt.row]; got != tt.want {
			t.Errorf("house %d line %d = %q, want %q", tt.house, tt.row, got, tt.want)
		}
	}
	for n := 1; n <= 12; n++ {
		for i, l := range tile(out, n) {
			if w := lipgloss.Width(l); w != houseWidth {
				t.Errorf("house %d line %d is %d wide, want %d: %q", n, i, w, houseWidth, l)
			}
		}
	}
}

func TestSquareTileASCII(t *testing.T) {
	out := Square(testSquare(), Options{Symbols: ASCII})
	if got, want := tile(out, 1)[2], "│Ta     Taurus      E│"; got != want {
		t.Errorf("house sign row = %q, want %q", got, want)
	}
	if got, want := tile(out, 1)[9], "│Li    Amissio      A│"; got != want {
		t.Errorf("figure row = %q, want %q", got, want)
	}
	if strings.ContainsAny(plain(out), "♈♉♊♋♌♍♎♏♐♑♒♓🜂🜁🜄🜃") {
		t.Error("ASCII square contains Unicode symbols")
	}
}

func TestSquareLegend(t *testing.T) {
	last := func(s string) string { lines := strings.Split(plain(s), "\n"); return lines[len(lines)-1] }
	if got := last(Square(testSquare(), Options{})); got != "Left: sign · Right: its element · 🜂 Fire · 🜁 Air · 🜄 Water · 🜃 Earth" {
		t.Errorf("legend = %q", got)
	}
	if got := last(Square(testSquare(), Options{Symbols: ASCII})); got != "Left: sign · Right: its element · F Fire · A Air · W Water · E Earth" {
		t.Errorf("ASCII legend = %q", got)
	}
}

func TestParseSymbols(t *testing.T) {
	for _, s := range []Symbols{Unicode, ASCII} {
		if got, err := ParseSymbols(strings.ToUpper(s.String())); err != nil || got != s {
			t.Errorf("ParseSymbols(%q) = %v, %v", s, got, err)
		}
	}
	if _, err := ParseSymbols("emoji"); err == nil {
		t.Error("ParseSymbols(emoji) succeeded")
	}
	if got := Symbols(9).String(); got != "Symbols(9)" {
		t.Errorf("Symbols(9).String() = %q", got)
	}
}

// The houses run counterclockwise around a 5×5 grid from the middle of the
// left side, with the corners and the centre empty.
func TestSquareLayout(t *testing.T) {
	want := map[int][2]int{ // house: row, column
		1: {2, 0}, 2: {3, 0}, // left side, middle and lower
		3: {4, 1}, 4: {4, 2}, 5: {4, 3}, // bottom, left to right
		6: {3, 4}, 7: {2, 4}, 8: {1, 4}, // right side, lower to upper
		9: {0, 3}, 10: {0, 2}, 11: {0, 1}, // top, right to left
		12: {1, 0}, // left side, upper
	}
	out := Square(testSquare(), Options{})
	lines := strings.Split(plain(out), "\n")
	const height = 12
	if got := len(lines); got != 5*height+1 {
		t.Fatalf("%d lines, want %d (5 rows of houses and a legend)", got, 5*height+1)
	}
	cellAt := func(r, c int) []string {
		var cell []string
		for _, l := range lines[r*height : (r+1)*height] {
			cell = append(cell, ansi.Cut(l, c*houseWidth, (c+1)*houseWidth))
		}
		return cell
	}

	occupied := map[[2]int]bool{}
	for house, rc := range want {
		occupied[rc] = true
		label := fmt.Sprintf("House %d ", house)
		if cell := cellAt(rc[0], rc[1]); !strings.Contains(cell[1], label) {
			t.Errorf("row %d column %d holds %q, want House %d", rc[0], rc[1], strings.TrimSpace(cell[1]), house)
		}
	}
	for r := range 5 {
		for c := range 5 {
			if occupied[[2]int{r, c}] {
				continue
			}
			for _, l := range cellAt(r, c) {
				if strings.TrimSpace(l) != "" {
					t.Errorf("row %d column %d should be empty, has %q", r, c, l)
					break
				}
			}
		}
	}
	if w := lipgloss.Width(strings.Join(lines[:5*height], "\n")); w != SquareWidth {
		t.Errorf("square is %d wide, want %d", w, SquareWidth)
	}
}
