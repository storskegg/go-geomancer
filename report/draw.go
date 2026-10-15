package report

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/tdewolff/canvas"

	"github.com/storskegg/go-geomancer/aspect"
	"github.com/storskegg/go-geomancer/figure"
)

// text draws a single line with its baseline at y.
func (p *page) text(f *canvas.FontFamily, size float64, col color.Color, x, y float64, s string, align canvas.TextAlign) {
	p.ctx.DrawText(x, y, canvas.NewTextLine(f.Face(size, col), s, align))
}

// wrap lays rich text out in lines no wider than width. A RichText can be
// laid out only once: a second ToText returns nothing.
func wrap(rt *canvas.RichText, width float64, align canvas.TextAlign) *canvas.Text {
	return rt.ToText(width, 0, align, canvas.Top, nil)
}

// paragraph draws laid-out text with its top at y, and returns the y of its
// bottom.
func (p *page) paragraph(t *canvas.Text, x, y float64) float64 {
	p.ctx.DrawText(x, y, t)
	return y - t.Bounds().H()
}

// header draws the title, the question, the system, significator and time,
// and an ornament, and returns the y below them.
func (rep *report) header(p *page) float64 {
	y := PageHeight - 34
	p.text(rep.f.title, 28, ink, PageWidth/2, y, "Geomantic Reading", canvas.Center)
	y -= 8

	if q := strings.TrimSpace(rep.r.Question); q != "" {
		t, width := rep.questionText("“" + q + "”")
		y = p.paragraph(t, (PageWidth-width)/2, y) - 3
	}

	meta := []string{rep.sys.Name()}
	if sig, ok := rep.opts.Content.Signification(rep.r.Significator); ok {
		title := sig.Title
		if l := rep.r.Matter.Label(); l != "" {
			title = l
		}
		meta = append(meta, fmt.Sprintf("Significator: House %d, %s", sig.House, title))
	}
	y -= 2
	p.text(rep.f.italic, 10.5, muted, PageWidth/2, y, strings.Join(meta, "  ·  "), canvas.Center)
	if c := rep.r.Circumstances; c != nil {
		y -= 5
		p.text(rep.f.italic, 10.5, muted, PageWidth/2, y,
			c.Time.Format("Monday, 2 January 2006, 3:04 pm MST")+"  ·  "+c.Place.Label(), canvas.Center)
		y -= 4.5
		p.text(rep.f.body, 8.5, muted, PageWidth/2, y,
			c.Place.Coordinates()+"  ·  "+c.TimeZone+" ("+c.UTCOffset()+")", canvas.Center)
	}
	y -= 5
	p.ornament(PageWidth/2, y)
	return y - 3
}

// questionLines is the most lines the question may take in the header, so
// the Square and its legends still fit on the first page.
const questionLines = 2

// questionText lays out the question in balanced lines, in the largest size
// from 13pt down to 10pt that fits questionLines; failing that, at 10pt,
// cut short with an ellipsis. It returns the text and the width it was laid
// out in, to centre it by.
func (rep *report) questionText(q string) (*canvas.Text, float64) {
	maxWidth := PageWidth - 2*(margin+10)
	lay := func(size float64, s string) (*canvas.Text, float64) {
		face := rep.f.italic.Face(size, ink)
		rt := canvas.NewRichText(face)
		rt.WriteString(s)
		width := balancedWidth(face.TextWidth(s), maxWidth)
		return wrap(rt, width, canvas.Center), width
	}
	for _, size := range []float64{13, 12, 11, 10} {
		if t, w := lay(size, q); t.Lines() <= questionLines {
			return t, w
		}
	}
	r := []rune(q)
	for n := len(r) - 1; n > 0; n-- {
		if t, w := lay(10, strings.TrimRight(string(r[:n]), " ,;:")+"…”"); t.Lines() <= questionLines {
			return t, w
		}
	}
	return lay(10, q)
}

// balancedWidth returns a width for text of natural width w that, wrapped
// within max, gives lines of similar length rather than a short last line.
func balancedWidth(w, maxWidth float64) float64 {
	lines := math.Ceil(w / maxWidth)
	if lines <= 1 {
		return maxWidth
	}
	return math.Min(maxWidth, w/lines*1.1)
}

// heading draws a section heading with its baseline at y and returns the y
// below it.
func (rep *report) heading(p *page, y float64, s string) float64 {
	p.text(rep.f.title, 17, ink, PageWidth/2, y, s, canvas.Center)
	p.ornament(PageWidth/2, y-4.5)
	return y - 7
}

// ornament draws a pair of rules with a diamond between them.
func (p *page) ornament(x, y float64) {
	ctx := p.ctx
	ctx.SetFillColor(canvas.Transparent)
	ctx.SetStrokeColor(rule)
	ctx.SetStrokeWidth(0.3)
	for _, s := range []float64{-1, 1} {
		ctx.MoveTo(x+s*4, y)
		ctx.LineTo(x+s*40, y)
		ctx.Stroke()
	}
	ctx.SetFillColor(rule)
	ctx.SetStrokeColor(canvas.Transparent)
	ctx.DrawPath(0, 0, polygon(pt{x - 2, y}, pt{x, y + 1.5}, pt{x + 2, y}, pt{x, y - 1.5}))
}

// dots draws a figure centred on x with its top row at y: one dot for an
// odd row, two for an even.
func (p *page) dots(f figure.Figure, col color.Color, x, y, r, dx, dy float64) {
	p.ctx.SetFillColor(col)
	p.ctx.SetStrokeColor(canvas.Transparent)
	for i, s := range f.Segments() {
		yy := y - float64(i)*dy
		if s == figure.Odd {
			p.ctx.DrawPath(x, yy, canvas.Circle(r))
		} else {
			p.ctx.DrawPath(x-dx/2, yy, canvas.Circle(r))
			p.ctx.DrawPath(x+dx/2, yy, canvas.Circle(r))
		}
	}
}

// aspectColour colours an aspect by its nature.
func aspectColour(a aspect.Aspect) color.RGBA {
	if a.Stressful() {
		return stressful
	}
	return easing
}

// glyph draws an aspect's symbol as a vector shape of size s centred on
// x, y, so it doesn't depend on a font having it.
func (p *page) glyph(a aspect.Aspect, x, y, s float64) {
	ctx := p.ctx
	col := aspectColour(a)
	ctx.SetFillColor(canvas.Transparent)
	ctx.SetStrokeColor(col)
	ctx.SetStrokeWidth(0.35)
	switch a {
	case aspect.Opposition:
		ctx.DrawPath(x-s*0.45, y-s*0.45, canvas.Circle(s*0.28))
		ctx.DrawPath(x+s*0.45, y+s*0.45, canvas.Circle(s*0.28))
		ctx.MoveTo(x-s*0.25, y-s*0.25)
		ctx.LineTo(x+s*0.25, y+s*0.25)
		ctx.Stroke()
	case aspect.Square:
		ctx.DrawPath(x-s/2, y-s/2, canvas.Rectangle(s, s))
	case aspect.Trine:
		ctx.DrawPath(0, 0, polygon(pt{x - s/2, y - s*0.4}, pt{x + s/2, y - s*0.4}, pt{x, y + s*0.48}))
	case aspect.Sextile:
		for k := range 3 {
			t := float64(k) * math.Pi / 3
			ctx.MoveTo(x-s/2*math.Cos(t), y-s/2*math.Sin(t))
			ctx.LineTo(x+s/2*math.Cos(t), y+s/2*math.Sin(t))
			ctx.Stroke()
		}
	}
}

// pt is a point in millimetres.
type pt struct{ x, y float64 }

func polygon(ps ...pt) *canvas.Path {
	path := &canvas.Path{}
	path.MoveTo(ps[0].x, ps[0].y)
	for _, q := range ps[1:] {
		path.LineTo(q.x, q.y)
	}
	path.Close()
	return path
}

// wrapName splits a figure's name onto two lines if it's wider than width.
func wrapName(face *canvas.FontFace, name string, width float64) []string {
	if face.TextWidth(name) <= width {
		return []string{name}
	}
	if i := strings.IndexByte(name, ' '); i > 0 {
		return []string{name[:i], name[i+1:]}
	}
	return []string{name}
}
