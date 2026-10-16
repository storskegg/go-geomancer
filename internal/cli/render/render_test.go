package render

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/storskegg/go-geomancer/aspect"
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
		"square": {Square(sq, 0, Options{}), SquareWidth},
		"ascii":  {Square(sq, 0, Options{Symbols: ASCII}), SquareWidth},
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
	// Golden Dawn gives Caput Draconis Virgo.
	for line := range strings.SplitSeq(plain(Figures(goldendawn.System)), "\n") {
		if strings.Contains(line, "Caput Draconis") && !strings.Contains(line, "│ Virgo") {
			t.Errorf("Caput Draconis row lacks its sign: %q", line)
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
	out := plain(Square(sq, 0, Options{}))
	for _, h := range sq {
		if !strings.Contains(out, fmt.Sprintf("House %d ", h.Number)) || !strings.Contains(out, h.Sign.String()) {
			t.Errorf("square lacks house %d or its sign %v", h.Number, h.Sign)
		}
	}

	// A square with no signs shows dashes.
	for i := range sq {
		sq[i].Sign = nil
	}
	if out := plain(Square(sq, 0, Options{})); strings.Count(out, "—") != 12 {
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
	if out := plain(Figure(figure.CaputDraconis, goldendawn.System, content.Entry{})); !strings.Contains(out, "Planet    North Node") || !strings.Contains(out, "Sign      Virgo") {
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
		start := col * (houseWidth + houseGap)
		t = append(t, ansi.Cut(l, start, start+houseWidth))
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
	out := Square(testSquare(), 0, Options{})
	tests := []struct {
		house int
		row   int // line within the tile; 0 is the top border
		want  string
	}{
		// The house's sign symbol and element flank the sign's name.
		{1, 2, "│♉   Taurus    🜃│"},
		// The figure's sign symbol and element flank the top row of its
		// dots (Amissio, 1212).
		{1, 4, "│♎      ●      🜁│"},
		// The next row leaves the right edge free for a further element,
		// and the dots stay aligned with the row above.
		{1, 5, "│      ●   ●     │"},
		// The name and position get the full width.
		{1, 9, "│    Amissio     │"},
		{1, 10, "│       M1       │"},
		// A figure without a sign has blank edges (Cauda Draconis, 1112).
		{2, 2, "│♌     Leo     🜂│"},
		{2, 4, "│        ●       │"},
		{2, 9, "│ Cauda Draconis │"},
		// A house without a sign shows a dash and blank edges.
		{3, 2, "│        —       │"},
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
	out := Square(testSquare(), 0, Options{Symbols: ASCII})
	if got, want := tile(out, 1)[2], "│Ta   Taurus    E│"; got != want {
		t.Errorf("house sign row = %q, want %q", got, want)
	}
	if got, want := tile(out, 1)[4], "│Li      ●      A│"; got != want {
		t.Errorf("figure sign row = %q, want %q", got, want)
	}
	if strings.ContainsAny(plain(out), "♈♉♊♋♌♍♎♏♐♑♒♓🜂🜁🜄🜃") {
		t.Error("ASCII square contains Unicode symbols")
	}
}

func TestSquareLegend(t *testing.T) {
	last := func(s string) string { lines := strings.Split(plain(s), "\n"); return lines[len(lines)-1] }
	if got := last(Square(testSquare(), 0, Options{})); got != "Left: sign · Right: its element · 🜂 Fire · 🜁 Air · 🜄 Water · 🜃 Earth" {
		t.Errorf("legend = %q", got)
	}
	if got := last(Square(testSquare(), 0, Options{Symbols: ASCII})); got != "Left: sign · Right: its element · F Fire · A Air · W Water · E Earth" {
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
	out := Square(testSquare(), 0, Options{})
	lines := strings.Split(plain(out), "\n")
	const height = 12
	if got := len(lines); got != 5*height+1 {
		t.Fatalf("%d lines, want %d (5 rows of houses and a legend)", got, 5*height+1)
	}
	cellAt := func(r, c int) []string {
		var cell []string
		for _, l := range lines[r*height : (r+1)*height] {
			start := c * (houseWidth + houseGap)
			cell = append(cell, ansi.Cut(l, start, start+houseWidth))
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
	if w := lipgloss.Width(strings.Join(lines[:5*height], "\n")); w != SquareWidth || SquareWidth != 94 {
		t.Errorf("square is %d wide (SquareWidth %d), want 94", w, SquareWidth)
	}
}

// The longest sign and figure names fit on one line, with room to spare.
func TestSquareLongNames(t *testing.T) {
	sag := zodiac.Sagittarius
	sq := testSquare()
	sq[0].Sign, sq[0].Figure = &sag, figure.CaputDraconis
	sq[1].Figure = figure.FortunaMinor
	out := Square(sq, 0, Options{})
	if got, want := tile(out, 1)[2], "│♐ Sagittarius 🜂│"; got != want {
		t.Errorf("sign row = %q, want %q", got, want)
	}
	if got, want := tile(out, 1)[9], "│ Caput Draconis │"; got != want {
		t.Errorf("name row = %q, want %q", got, want)
	}
	if got, want := tile(out, 2)[9], "│ Fortuna Minor  │"; got != want {
		t.Errorf("name row = %q, want %q", got, want)
	}
}

func TestSquareSignificator(t *testing.T) {
	out := Square(testSquare(), 7, Options{})
	if got := tile(out, 7)[0]; !strings.HasPrefix(got, "┏") {
		t.Errorf("significator top border = %q, want heavy", got)
	}
	for _, n := range []int{1, 6, 8} {
		if got := tile(out, n)[0]; !strings.HasPrefix(got, "╭") {
			t.Errorf("house %d top border = %q, want rounded", n, got)
		}
	}
	lines := strings.Split(plain(out), "\n")
	legend := lines[len(lines)-2:]
	if want := "Heavy border: significator · ☍ Opposition, □ Square: stressful · △ Trine, ⚹ Sextile: easing"; legend[0] != want {
		t.Errorf("legend line = %q, want %q", legend[0], want)
	}
	if want := "↻ Dexter (clockwise), the stronger of a pair · ↺ Sinister"; legend[1] != want {
		t.Errorf("legend line = %q, want %q", legend[1], want)
	}
	if w := lipgloss.Width(out); w > SquareWidth {
		t.Errorf("square with significator is %d wide", w)
	}

	// Without one, no house is heavy or labelled, and the legend has one
	// line.
	if out := plain(Square(testSquare(), 0, Options{})); strings.Contains(out, "┏") || strings.Contains(out, "Heavy") || strings.Contains(out, "Square ↻") {
		t.Error("square without a significator marks one")
	}
}

// With the significator in House 10, the line under each house's sign
// names its relationship to it.
func TestSquareAspectLabels(t *testing.T) {
	const labelRow = 3 // border, number, sign, label
	tests := []struct {
		opts   Options
		labels map[int]string
	}{
		{Options{}, map[int]string{
			10: "  Significator  ",
			4:  "  ☍ Opposition  ",
			7:  "   □ Square ↻   ", 1: "   □ Square ↺   ",
			6: "   △ Trine ↻    ", 2: "   △ Trine ↺    ",
			8: "  ⚹ Sextile ↻   ", 12: "  ⚹ Sextile ↺   ",
			3: "", 5: "", 9: "", 11: "",
		}},
		{Options{Symbols: ASCII}, map[int]string{
			4: "   Opposition   ",
			7: "   Square dex   ", 1: "   Square sin   ",
			12: "  Sextile sin   ",
		}},
	}
	for _, tt := range tests {
		out := Square(testSquare(), 10, tt.opts)
		for house, want := range tt.labels {
			got := tile(out, house)[labelRow]
			got = strings.TrimSuffix(strings.TrimPrefix(got, "│"), "│")
			got = strings.TrimSuffix(strings.TrimPrefix(got, "┃"), "┃")
			if want == "" {
				if strings.TrimSpace(got) != "" {
					t.Errorf("%v: house %d label = %q, want none", tt.opts.Symbols, house, got)
				}
				continue
			}
			if got != want {
				t.Errorf("%v: house %d label = %q, want %q", tt.opts.Symbols, house, got, want)
			}
		}
		if w := lipgloss.Width(out); w > SquareWidth {
			t.Errorf("%v: square is %d wide", tt.opts.Symbols, w)
		}
	}
}

func TestAspectsLine(t *testing.T) {
	houses, err := aspect.Houses(10)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := plain(Aspects(houses, Options{})), "Aspects: ☍ Opposition 4 · □ Square 7 ↻, 1 ↺ · △ Trine 6 ↻, 2 ↺ · ⚹ Sextile 8 ↻, 12 ↺"; got != want {
		t.Errorf("Aspects =\n%q\nwant\n%q", got, want)
	}
	if got, want := plain(Aspects(houses, Options{Symbols: ASCII})), "Aspects: Opposition 4 · Square 7 dex, 1 sin · Trine 6 dex, 2 sin · Sextile 8 dex, 12 sin"; got != want {
		t.Errorf("ASCII Aspects =\n%q\nwant\n%q", got, want)
	}
}

func TestSideMark(t *testing.T) {
	if got := (Options{}).sideMark(aspect.Neither); got != "" {
		t.Errorf("sideMark(Neither) = %q", got)
	}
}

func TestSignificatorLine(t *testing.T) {
	sig := content.Signification{House: 7, Title: "Partnership", Text: "marriage, partners"}
	for _, tt := range []struct {
		matter *chart.Matter
		want   string
	}{
		{nil, "Significator: House 7 · Partnership — marriage, partners"},
		{&chart.Matter{}, "Significator: House 7 · Partnership — marriage, partners"},
		{&chart.Matter{Topics: []string{"thieves"}}, "Significator: House 7 · Thieves — marriage, partners"},
	} {
		if got := plain(Significator(sig, tt.matter)); got != tt.want {
			t.Errorf("Significator(%v) = %q, want %q", tt.matter, got, tt.want)
		}
	}
}

func TestMeanings(t *testing.T) {
	r, err := chart.NewReading([4]figure.Figure{figure.Amissio, figure.FortunaMinor, figure.Via, figure.Populus}, goldendawn.System)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.SetSignificator(10); err != nil {
		t.Fatal(err)
	}
	out := plain(Meanings(r, content.Default(), Options{}))
	blocks := strings.Split(out, "\n\n")
	if len(blocks) != 8 {
		t.Fatalf("%d houses, want 8:\n%s", len(blocks), out)
	}
	heads := []string{
		"House 1 · Querent · □ Square ↺ · Self",
		"House 10 · Significator · Honour",
		"House 4 · ☍ Opposition · Home",
		"House 7 · □ Square ↻ · Partnership",
		"House 6 · △ Trine ↻ · Illness",
		"House 2 · △ Trine ↺ · Wealth",
		"House 8 · ⚹ Sextile ↻ · Death",
		"House 12 · ⚹ Sextile ↺ · Sorrow",
	}
	for i, want := range heads {
		if got := strings.SplitN(blocks[i], "\n", 2)[0]; got != want {
			t.Errorf("block %d heading = %q, want %q", i, got, want)
		}
	}
	// Each house's figure and its meaning in that house.
	for i, f := range r.Focus() {
		fig := r.Square[f.House-1].Figure
		body := strings.Join(strings.Fields(strings.SplitN(blocks[i], "\n", 2)[1]), " ")
		want := fig.Name() + " — " + content.Default().For(fig).House(f.House)
		if body != want {
			t.Errorf("house %d body = %q, want %q", f.House, body, want)
		}
	}
	if w := lipgloss.Width(Meanings(r, content.Default(), Options{})); w > Width {
		t.Errorf("meanings are %d wide, want at most %d", w, Width)
	}
}

func TestMeaningsUnwritten(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"locale": "xx", "significations": [`)
	for h := 1; h <= 12; h++ {
		if h > 1 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"house": %d, "title": "T%d"}`, h, h)
	}
	b.WriteString(`], "figures": [`)
	for i, f := range figure.All() {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"figure": %q}`, f.Name())
	}
	b.WriteString("]}")
	c, err := content.Parse([]byte(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	r, _ := chart.NewReading([4]figure.Figure{}, goldendawn.System)
	out := plain(Meanings(r, c, Options{}))
	if got := strings.Count(out, "Not yet written."); got != 12 {
		t.Errorf("%d placeholders, want 12:\n%s", got, out)
	}
	if !strings.HasPrefix(out, "House 1 · Querent · T1\n  Populus — ") {
		t.Errorf("first block:\n%.80s", out)
	}
}
