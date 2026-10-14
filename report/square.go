package report

import (
	"fmt"
	"math"
	"strings"

	"github.com/tdewolff/canvas"

	"github.com/storskegg/go-geomancer/aspect"
	"github.com/storskegg/go-geomancer/zodiac"
)

// squareSize is the width of the square chart, and squareHalf half of it,
// in millimetres.
const (
	squareSize = 150.0
	squareHalf = squareSize / 2
)

// houseTriangle returns house n's triangle (1–12) in the traditional square
// chart, centred on the origin.
//
// The chart is an outer square; a centre square half its width; a diamond
// joining the midpoints of the outer sides, whose edges pass through the
// centre square's corners; and diagonals from each outer corner to the
// centre square's. That makes twelve triangles of equal area. They run
// counterclockwise from House 1 on the left; the angular houses (1, 4, 7,
// 10) point outward from the centre square's sides.
func houseTriangle(n int) [3]pt {
	h := squareHalf
	q := h / 2
	l, r, t, b := pt{-h, 0}, pt{h, 0}, pt{0, h}, pt{0, -h}
	tl, bl, br, tr := pt{-h, h}, pt{-h, -h}, pt{h, -h}, pt{h, h}
	itl, ibl, ibr, itr := pt{-q, q}, pt{-q, -q}, pt{q, -q}, pt{q, q}
	return [12][3]pt{
		{l, itl, ibl}, {l, bl, ibl}, {bl, b, ibl},
		{b, ibl, ibr}, {b, br, ibr}, {br, r, ibr},
		{r, ibr, itr}, {r, tr, itr}, {tr, t, itr},
		{t, itr, itl}, {t, tl, itl}, {tl, l, itl},
	}[n-1]
}

func dist(a, b pt) float64 { return math.Hypot(a.x-b.x, a.y-b.y) }

// incentre returns the centre of a triangle's inscribed circle, where a
// house's contents sit.
func incentre(t [3]pt) pt {
	a, b, c := dist(t[1], t[2]), dist(t[0], t[2]), dist(t[0], t[1])
	s := a + b + c
	return pt{(a*t[0].x + b*t[1].x + c*t[2].x) / s, (a*t[0].y + b*t[1].y + c*t[2].y) / s}
}

// aspectAnchor returns where house n's aspect line ends, on a circle of
// radius rad in the centre square: House 1 on the left, then
// counterclockwise like the houses.
func aspectAnchor(n int, rad float64) pt {
	a := float64(180+(n-1)*30) * math.Pi / 180
	return pt{rad * math.Cos(a), rad * math.Sin(a)}
}

// square draws the Astrological Square as a traditional square chart below
// top and returns the y of its bottom edge.
func (rep *report) square(p *page, top float64) float64 {
	h := squareHalf
	cx, cy := PageWidth/2, top-h
	ctx := p.ctx
	ctx.Push()
	defer ctx.Pop()
	ctx.Translate(cx, cy)

	if rep.r.Significator != 0 {
		t := houseTriangle(rep.r.Significator)
		ctx.SetFillColor(sigShade)
		ctx.SetStrokeColor(canvas.Transparent)
		ctx.DrawPath(0, 0, polygon(t[:]...))
	}

	ctx.SetFillColor(canvas.Transparent)
	ctx.SetStrokeColor(ink)
	ctx.SetStrokeWidth(lineWidth * 1.6)
	ctx.DrawPath(-h, -h, canvas.Rectangle(squareSize, squareSize))
	ctx.SetStrokeWidth(lineWidth)
	ctx.DrawPath(-h/2, -h/2, canvas.Rectangle(h, h))
	ctx.DrawPath(0, 0, polygon(pt{-h, 0}, pt{0, h}, pt{h, 0}, pt{0, -h}))
	for _, s := range [][2]float64{{-1, 1}, {-1, -1}, {1, -1}, {1, 1}} {
		ctx.MoveTo(s[0]*h, s[1]*h)
		ctx.LineTo(s[0]*h/2, s[1]*h/2)
		ctx.Stroke()
	}

	f := rep.f
	for _, house := range rep.r.Square {
		c := incentre(houseTriangle(house.Number))
		col := ink
		if house.Number == rep.r.Significator {
			col = accent
		}
		p.text(f.body, 7.5, muted, c.x, c.y+11.5, fmt.Sprint(house.Number), canvas.Center)
		if house.Sign != nil {
			p.text(f.symbols, 10, col, c.x-2.4, c.y+6.2, house.Sign.Symbol(), canvas.Center)
			p.text(f.symbols, 8.5, elementColours[house.Sign.Element()], c.x+3.2, c.y+6.2, house.Sign.Element().Symbol(), canvas.Center)
		}
		p.dots(house.Figure, col, c.x, c.y+2.5, 0.85, 2.4, 2.5)
		lines := wrapName(f.italic.Face(7.5), house.Figure.Name(), 26)
		for i, l := range lines {
			p.text(f.italic, 7.5, col, c.x, c.y-9.6-float64(i)*3, l, canvas.Center)
		}
	}

	rep.aspectLines(p, h)
	return cy - h
}

// aspectLines draws, across the centre square, a line from the significator
// to each house aspecting it: solid for Dexter, dashed for Sinister, each
// ending in the aspect's symbol.
func (rep *report) aspectLines(p *page, h float64) {
	if rep.r.Significator == 0 {
		return
	}
	ctx := p.ctx
	rad := h / 2 * 0.78
	s := aspectAnchor(rep.r.Significator, rad)
	for _, a := range rep.r.Aspects {
		e := aspectAnchor(a.House, rad)
		ctx.SetFillColor(canvas.Transparent)
		ctx.SetStrokeColor(aspectColour(a.Aspect))
		ctx.SetStrokeWidth(0.5)
		if a.Side == aspect.Sinister {
			ctx.SetDashes(0, 1.6, 1.0)
		}
		ctx.MoveTo(s.x, s.y)
		ctx.LineTo(e.x, e.y)
		ctx.Stroke()
		ctx.SetDashes(0)
		p.glyph(a.Aspect, e.x, e.y, 2.2)
	}
	ctx.SetFillColor(accent)
	ctx.SetStrokeColor(canvas.Transparent)
	ctx.DrawPath(s.x, s.y, canvas.Circle(1.1))
}

// squareLegend explains the elements and, if there's a significator, the
// aspect lines, centred below the chart with its first baseline at y.
func (rep *report) squareLegend(p *page, y float64) {
	f := rep.f
	ctx := p.ctx

	// Elements: symbol and name, spaced evenly.
	x := PageWidth/2 - 60
	for _, e := range zodiac.Elements() {
		p.text(f.symbols, 9, elementColours[e], x, y, e.Symbol(), canvas.Left)
		p.text(f.body, 9, muted, x+4.5, y, e.String(), canvas.Left)
		x += 30
	}

	if rep.r.Significator == 0 {
		return
	}
	y -= 7
	x = PageWidth/2 - 84
	for _, a := range aspect.All() {
		p.glyph(a, x, y+1.1, 2.6)
		p.text(f.body, 9, muted, x+3, y, a.String(), canvas.Left)
		x += 26
	}
	ctx.SetFillColor(canvas.Transparent)
	ctx.SetStrokeColor(muted)
	ctx.SetStrokeWidth(0.5)
	for i, label := range []string{"Dexter", "Sinister"} {
		if i == 1 {
			ctx.SetDashes(0, 1.6, 1.0)
		}
		ctx.MoveTo(x, y+1.1)
		ctx.LineTo(x+7, y+1.1)
		ctx.Stroke()
		ctx.SetDashes(0)
		p.text(f.body, 9, muted, x+9, y, label, canvas.Left)
		x += 26
	}
	y -= 6
	p.text(f.italic, 8.5, muted, PageWidth/2, y, strings.Join([]string{
		"Shaded: the significator", "red: stressful", "green: easing", "Dexter is the stronger of a pair",
	}, "  ·  "), canvas.Center)
}
