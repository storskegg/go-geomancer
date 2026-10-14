package report

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"strings"
	"testing"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/system/goldendawn"
)

func area(t [3]pt) float64 {
	return math.Abs((t[1].x-t[0].x)*(t[2].y-t[0].y)-(t[2].x-t[0].x)*(t[1].y-t[0].y)) / 2
}

// inside reports whether q is inside triangle t.
func inside(t [3]pt, q pt) bool {
	sub := func(a, b, c pt) float64 { return (b.x-a.x)*(c.y-a.y) - (c.x-a.x)*(b.y-a.y) }
	d1, d2, d3 := sub(t[0], t[1], q), sub(t[1], t[2], q), sub(t[2], t[0], q)
	neg := d1 < 0 || d2 < 0 || d3 < 0
	pos := d1 > 0 || d2 > 0 || d3 > 0
	return !neg || !pos
}

// The twelve houses are equal triangles that tile the square around the
// centre square, counterclockwise from House 1 on the left.
func TestHouseTriangles(t *testing.T) {
	const h = squareHalf
	total := 0.0
	prev := -1.0
	for n := 1; n <= 12; n++ {
		tri := houseTriangle(n)
		a := area(tri)
		if math.Abs(a-h*h/4) > 1e-9 {
			t.Errorf("house %d area = %v, want %v", n, a, h*h/4)
		}
		total += a
		c := incentre(tri)
		if !inside(tri, c) {
			t.Errorf("house %d incentre %v is outside its triangle", n, c)
		}
		if math.Abs(c.x) < h/2 && math.Abs(c.y) < h/2 {
			t.Errorf("house %d incentre %v is in the centre square", n, c)
		}
		// Angles increase counterclockwise from House 1 at 180°.
		ang := math.Mod(math.Atan2(c.y, c.x)*180/math.Pi-180+720, 360)
		if ang < prev {
			t.Errorf("house %d at %.0f° comes before house %d at %.0f°", n, ang, n-1, prev)
		}
		prev = ang
	}
	if want := (2*h)*(2*h) - h*h; math.Abs(total-want) > 1e-6 {
		t.Errorf("houses cover %v, want %v (the square less the centre)", total, want)
	}

	// The angular houses point outward from the centre square's sides.
	for n, dir := range map[int]pt{1: {-1, 0}, 4: {0, -1}, 7: {1, 0}, 10: {0, 1}} {
		c := incentre(houseTriangle(n))
		if c.x*dir.x+c.y*dir.y < h/2 || math.Abs(c.x*dir.y-c.y*dir.x) > 1e-9 {
			t.Errorf("house %d incentre %v doesn't lie along %v", n, c, dir)
		}
	}
}

func TestAspectAnchor(t *testing.T) {
	for n, want := range map[int]pt{1: {-10, 0}, 4: {0, -10}, 7: {10, 0}, 10: {0, 10}} {
		got := aspectAnchor(n, 10)
		if math.Abs(got.x-want.x) > 1e-9 || math.Abs(got.y-want.y) > 1e-9 {
			t.Errorf("aspectAnchor(%d) = %v, want %v", n, got, want)
		}
	}
}

func TestShieldHalfWidth(t *testing.T) {
	if shieldHalfWidth(0) != 1 || shieldHalfWidth(0.5) != 1 {
		t.Error("the Shield's sides aren't straight above the curve")
	}
	prev := 1.0
	for d := 0.56; d <= 1; d += 0.02 {
		w := shieldHalfWidth(d)
		if w > prev+1e-9 || w < 0 {
			t.Errorf("half-width %v at depth %.2f after %v", w, d, prev)
		}
		prev = w
	}
	if w := shieldHalfWidth(1); w > 0.01 {
		t.Errorf("half-width at the point = %v, want 0", w)
	}
}

func TestWrapName(t *testing.T) {
	f, err := loadFonts()
	if err != nil {
		t.Fatal(err)
	}
	face := f.italic.Face(9)
	if got := wrapName(face, "Via", 10); len(got) != 1 {
		t.Errorf("wrapName(Via) = %q", got)
	}
	if got := wrapName(face, "Caput Draconis", 10); len(got) != 2 || got[0] != "Caput" || got[1] != "Draconis" {
		t.Errorf("wrapName(Caput Draconis) = %q", got)
	}
	if got := wrapName(face, "Conjunctio", 1); len(got) != 1 {
		t.Errorf("one word can't wrap: %q", got)
	}
}

func TestBalancedWidth(t *testing.T) {
	if got := balancedWidth(50, 100); got != 100 {
		t.Errorf("one line: %v, want the full width", got)
	}
	// 150 wide in 100 is two lines of about 75.
	if got := balancedWidth(150, 100); got < 75 || got > 100 {
		t.Errorf("two lines: %v", got)
	}
}

func reading(t *testing.T, sig int) chart.Reading {
	t.Helper()
	r, err := chart.NewReading([4]figure.Figure{figure.Amissio, figure.FortunaMinor, figure.Via, figure.Populus}, goldendawn.System)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.SetSignificator(sig); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestPages(t *testing.T) {
	for _, sig := range []int{0, 10} {
		pages, err := Pages(reading(t, sig), Options{})
		if err != nil {
			t.Fatal(err)
		}
		if len(pages) != 3 {
			t.Errorf("significator %d: %d pages, want 3 (Square, Shield, Reading)", sig, len(pages))
		}
		for i, p := range pages {
			if p.W != PageWidth || p.H != PageHeight {
				t.Errorf("page %d is %v × %v mm, want US Letter", i+1, p.W, p.H)
			}
		}
	}
	r := reading(t, 0)
	r.System = "nope"
	if _, err := Pages(r, Options{}); err == nil {
		t.Error("unknown system succeeded")
	}
}

// A long Reading continues onto further pages.
func TestPagination(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"locale": "en", "significations": [`)
	for h := 1; h <= 12; h++ {
		if h > 1 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"house": %d, "title": "T%d"}`, h, h)
	}
	b.WriteString(`], "figures": [`)
	long := strings.Repeat("A long meaning that goes on and on. ", 30)
	for i, f := range figure.All() {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"figure": %q, "houses": {`, f.Name())
		for h := 1; h <= 12; h++ {
			if h > 1 {
				b.WriteString(",")
			}
			fmt.Fprintf(&b, `"%d": %q`, h, long)
		}
		b.WriteString("}}")
	}
	b.WriteString("]}")
	c, err := content.Parse([]byte(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	pages, err := Pages(reading(t, 0), Options{Content: c})
	if err != nil {
		t.Fatal(err)
	}
	if len(pages) < 5 {
		t.Errorf("%d pages for twelve long meanings, want the Reading to run over several", len(pages))
	}
}

func TestWritePDF(t *testing.T) {
	var buf bytes.Buffer
	r := reading(t, 10)
	r.Question = "Will it work?"
	if err := WritePDF(&buf, r, Options{}); err != nil {
		t.Fatal(err)
	}
	out := buf.Bytes()
	if !bytes.HasPrefix(out, []byte("%PDF-")) || !bytes.HasSuffix(bytes.TrimSpace(out), []byte("%%EOF")) {
		t.Errorf("not a complete PDF: starts %q, ends %q", out[:8], out[len(out)-8:])
	}
	if !bytes.Contains(out, []byte("/Count 3")) {
		t.Error("PDF doesn't have 3 pages")
	}
	if n := bytes.Count(out, []byte("/FontFile2")); n < 4 {
		t.Errorf("%d embedded fonts, want the report's fonts embedded", n)
	}
}

// at samples a page's colour at x, y millimetres.
func at(img interface{ At(x, y int) color.Color }, x, y float64, dpmm float64) color.RGBA {
	c := color.RGBAModel.Convert(img.At(int(x*dpmm), int((PageHeight-y)*dpmm))).(color.RGBA)
	return c
}

func near(a, b color.RGBA) bool {
	d := func(x, y uint8) int { return int(x) - int(y) }
	sq := func(v int) int { return v * v }
	return sq(d(a.R, b.R))+sq(d(a.G, b.G))+sq(d(a.B, b.B)) < 12*12
}

// The significator's house is shaded and no other is.
func TestSignificatorShaded(t *testing.T) {
	const dpmm = 2.0
	for _, sig := range []int{0, 4, 10} {
		pages, err := Pages(reading(t, sig), Options{})
		if err != nil {
			t.Fatal(err)
		}
		img := rasterizer.Draw(pages[0], canvas.DPMM(dpmm), canvas.DefaultColorSpace)

		// Find the Square: below the header, centred. Sample each house
		// two-thirds of the way from its incentre to its outermost corner,
		// clear of its contents.
		r := reading(t, sig)
		rep := &report{r: r, sys: goldendawn.System, opts: Options{Content: content.Default()}}
		rep.f, _ = loadFonts()
		top := rep.header(&page{c: canvas.New(PageWidth, PageHeight), ctx: canvas.NewContext(canvas.New(PageWidth, PageHeight))}) - 6
		h := squareHalf
		cx, cy := PageWidth/2, top-h
		for n := 1; n <= 12; n++ {
			tri := houseTriangle(n)
			c := incentre(tri)
			far := tri[0]
			for _, v := range tri {
				if math.Hypot(v.x, v.y) > math.Hypot(far.x, far.y) {
					far = v
				}
			}
			x, y := cx+c.x+(far.x-c.x)*0.6, cy+c.y+(far.y-c.y)*0.6
			got := at(img, x, y, dpmm)
			want := parchment
			if n == sig {
				want = sigShade
			}
			if !near(got, want) {
				t.Errorf("significator %d: house %d at (%.0f, %.0f) is %v, want %v", sig, n, x, y, got, want)
			}
		}
	}
}
