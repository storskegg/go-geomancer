package report

import (
	"image/color"

	"github.com/tdewolff/canvas"

	"github.com/storskegg/go-geomancer/position"
)

// The Shield's size, in millimetres.
const (
	shieldWidth  = 172.0
	shieldHeight = 172.0
)

// A Shield cell's contents, top to bottom: the position, the figure's dots
// and its name, in millimetres.
const (
	cellDotR      = 1.2
	cellDotStep   = 3.3
	cellContent   = 26.0 // height of the contents, for centring
	cellNameGap   = 7.5  // from the last row of dots to the name's baseline
	cellLabelDrop = 5.5  // from the label's baseline to the first row of dots
)

// shieldCurve is the right half of the Shield's lower edge: a cubic Bézier
// from the bottom of the straight right side to the point, in units of the
// half-width (x) and height (y, downward from the top).
var shieldCurve = [4]pt{{1, 0.55}, {1, 0.86}, {0.32, 0.97}, {0, 1}}

func bezier(c [4]pt, t float64) pt {
	u := 1 - t
	a, b, d, e := u*u*u, 3*u*u*t, 3*u*t*t, t*t*t
	return pt{a*c[0].x + b*c[1].x + d*c[2].x + e*c[3].x, a*c[0].y + b*c[1].y + d*c[2].y + e*c[3].y}
}

// shieldHalfWidth returns the Shield's half-width, as a fraction of its
// greatest, at depth d (0 at the top, 1 at the point).
func shieldHalfWidth(d float64) float64 {
	if d <= shieldCurve[0].y {
		return 1
	}
	lo, hi := 0.0, 1.0 // the curve's depth increases with t
	for range 40 {
		mid := (lo + hi) / 2
		if bezier(shieldCurve, mid).y < d {
			lo = mid
		} else {
			hi = mid
		}
	}
	return bezier(shieldCurve, lo).x
}

// shieldRows are the Shield's rows, top to bottom, each listed left to right:
// the Mothers and Daughters reading right to left with M1 on the right, the
// Nieces, the Witnesses, and the Judge.
var shieldRows = [][]position.Position{
	{position.D4, position.D3, position.D2, position.D1, position.M4, position.M3, position.M2, position.M1},
	{position.N4, position.N3, position.N2, position.N1},
	{position.WL, position.WR},
	{position.J},
}

// shield draws the Shield below top, with the Judgement beside its point,
// and returns the y below its legend.
func (rep *report) shield(p *page, top float64) float64 {
	a, H := shieldWidth/2, shieldHeight
	cx := PageWidth / 2
	at := func(x, d float64) pt { return pt{cx + x*a, top - d*H} } // half-width units, depth

	outline := &canvas.Path{}
	start := at(-1, 0)
	outline.MoveTo(start.x, start.y)
	for _, q := range []pt{at(1, 0), at(1, shieldCurve[0].y)} {
		outline.LineTo(q.x, q.y)
	}
	c1, c2, c3 := at(shieldCurve[1].x, shieldCurve[1].y), at(shieldCurve[2].x, shieldCurve[2].y), at(0, 1)
	outline.CubeTo(c1.x, c1.y, c2.x, c2.y, c3.x, c3.y)
	c1, c2, c3 = at(-shieldCurve[2].x, shieldCurve[2].y), at(-shieldCurve[1].x, shieldCurve[1].y), at(-1, shieldCurve[0].y)
	outline.CubeTo(c1.x, c1.y, c2.x, c2.y, c3.x, c3.y)
	outline.Close()

	ctx := p.ctx
	ctx.SetFillColor(canvas.Transparent)
	ctx.SetStrokeColor(ink)
	ctx.SetStrokeWidth(lineWidth * 1.6)
	ctx.DrawPath(0, 0, outline)

	// Rules between the rows, each as wide as the Shield at its depth, and
	// between the figures in each row.
	ctx.SetStrokeWidth(lineWidth)
	rowDepth := 1.0 / float64(len(shieldRows))
	for i := 1; i < len(shieldRows); i++ {
		d := float64(i) * rowDepth
		w := shieldHalfWidth(d)
		l, r := at(-w, d), at(w, d)
		ctx.MoveTo(l.x, l.y)
		ctx.LineTo(r.x, r.y)
		ctx.Stroke()
	}
	for i, row := range shieldRows[:len(shieldRows)-1] {
		for j := 1; j < len(row); j++ {
			x := -1 + 2*float64(j)/float64(len(row))
			t, b := at(x, float64(i)*rowDepth), at(x, float64(i+1)*rowDepth)
			ctx.MoveTo(t.x, t.y)
			ctx.LineTo(b.x, b.y)
			ctx.Stroke()
		}
	}

	f := rep.f
	s := rep.r.Shield
	// cell draws a position's label, figure and name, with the label's
	// baseline at y.
	cell := func(pos position.Position, label string, x, y, width float64, col color.Color) {
		p.text(f.body, 8, muted, x, y, label, canvas.Center)
		dotsTop := y - cellLabelDrop
		p.dots(s.At(pos), col, x, dotsTop, cellDotR, cellDotStep, cellDotStep)
		nameY := dotsTop - 3*cellDotStep - cellNameGap
		for k, l := range wrapName(f.italic.Face(9), s.At(pos).Name(), width-3) {
			p.text(f.italic, 9, col, x, nameY-float64(k)*3.6, l, canvas.Center)
		}
	}
	rowH := rowDepth * H
	for i, row := range shieldRows {
		// Centre the contents in the row; the Judge's sits higher, as the
		// point narrows below it.
		y := top - float64(i)*rowH - (rowH-cellContent)/2 - 3
		if pos := row[0]; pos == position.J {
			y = top - float64(i)*rowH - 7
		}
		for j, pos := range row {
			x := -1 + (2*float64(j)+1)/float64(len(row))
			col := ink
			if pos == position.J {
				col = accent
			}
			c := at(x, 0)
			cell(pos, pos.String(), c.x, y, 2*a/float64(len(row)), col)
		}
	}

	// The Judgement, outside the point, to its right.
	jm := at(0.9, 0.86)
	cell(position.JM, "Judgement", jm.x, jm.y, 30, ink)

	nieces := "Nieces"
	if rep.opts.Nephews {
		nieces = "Nephews"
	}
	y := top - H - 7
	p.text(f.italic, 8.5, muted, PageWidth/2, y,
		"M Mothers  ·  D Daughters  ·  N "+nieces+"  ·  W Witnesses  ·  J Judge  ·  JM Judgement", canvas.Center)
	return y
}
