package report

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/tdewolff/canvas"

	"github.com/storskegg/go-geomancer/aspect"
	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/sky"
	"github.com/storskegg/go-geomancer/zodiac"
)

// heavensLine summarizes the sky for the header: "Day of Jupiter · Hour of
// Saturn · Crescent Moon in Sagittarius".
func heavensLine(s *sky.Sky) string {
	parts := []string{"Day " + of(s.Day.Ruler)}
	if h := s.Day.Hour; h != nil {
		parts = append(parts, "Hour "+of(h.Ruler))
	}
	parts = append(parts, fmt.Sprintf("%s Moon in %s", s.Moon.Phase, s.Body(zodiac.Moon).Sign))
	return strings.Join(parts, "  ·  ")
}

// of names a planet after "of": "of Jupiter", but "of the Sun".
func of(p zodiac.Planet) string {
	if p == zodiac.Sun || p == zodiac.Moon {
		return "of the " + p.String()
	}
	return "of " + p.String()
}

func clock(t time.Time) string { return t.Format("3:04 pm") }

// heavens draws the astrology of the moment on a page of its own: the day,
// hour and Moon, the planets and nodes, their aspects, and the lunations;
// then the timing if it fits. It returns the page and the y below what it
// drew, and whether the timing fitted.
func (rep *report) heavens(s *sky.Sky) (*page, float64, bool) {
	p := rep.newPage()
	f := rep.f
	y := rep.heading(p, PageHeight-34, "The Heavens") - 7

	for _, line := range heavensSummary(s) {
		p.text(f.italic, 10.5, muted, PageWidth/2, y, line, canvas.Center)
		y -= 5
	}

	y = rep.bodies(p, s, y-7)
	y = rep.skyAspects(p, s, y-9)
	y = rep.lunations(p, s.Moon, y-9)
	if y-12-rep.timingHeight(rep.r.Timing) < bottom {
		return p, y, false
	}
	return p, rep.timing(p, rep.r.Timing, y-12), true
}

// Timing layout, in mm.
const (
	markWidth, aboutWidth = 5.0, 13.0
	timingWidth           = PageWidth - 2*tableLeft - markWidth - aboutWidth
	timingHeading         = 8.0
	timingGap             = 2.0
)

// timingTexts lays out the judgements' texts.
func (rep *report) timingTexts(js []chart.Judgement) []*canvas.Text {
	texts := make([]*canvas.Text, len(js))
	for i, j := range js {
		rt := canvas.NewRichText(rep.f.body.Face(10.5, ink))
		rt.WriteString(j.Text)
		texts[i] = wrap(rt, timingWidth, canvas.Left)
	}
	return texts
}

// timingHeight is the height timing would take, 0 for no judgements.
func (rep *report) timingHeight(js []chart.Judgement) float64 {
	if len(js) == 0 {
		return 0
	}
	h := timingHeading
	for _, t := range rep.timingTexts(js) {
		h += t.Bounds().H() + timingGap
	}
	return h
}

// timing draws the judgements of the moment against the matter from y
// down, each marked by whether it helps (+), hinders (−), is mixed (±) or
// neither (·), and returns the y below them.
func (rep *report) timing(p *page, js []chart.Judgement, y float64) float64 {
	if len(js) == 0 {
		return y
	}
	f := rep.f
	texts := rep.timingTexts(js)
	p.text(f.title, 13, ink, PageWidth/2, y, "Timing", canvas.Center)
	y -= timingHeading
	for i, j := range js {
		mark, col := effectMark(j.Effect)
		p.text(f.bold, 11, col, tableLeft+1.5, y-3.4, mark, canvas.Center)
		p.text(f.italic, 10.5, muted, tableLeft+markWidth, y-3.4, j.About, canvas.Left)
		y = p.paragraph(texts[i], tableLeft+markWidth+aboutWidth, y) - timingGap
	}
	return y
}

func effectMark(e chart.Effect) (string, color.RGBA) {
	switch e {
	case chart.Helps:
		return "+", easing
	case chart.Hinders:
		return "−", stressful
	case chart.Mixed:
		return "±", mixed
	case chart.Neutral:
	}
	return "·", muted
}

// heavensSummary describes the day, hour and Moon in a few lines.
func heavensSummary(s *sky.Sky) []string {
	d := s.Day
	var lines []string
	if h := d.Hour; h != nil {
		part := "day"
		if h.Night {
			part = "night"
		}
		lines = append(lines,
			fmt.Sprintf("The day %s, from sunrise on %s at %s", of(d.Ruler), d.Sunrise.Format("Monday"), clock(d.Sunrise)),
			fmt.Sprintf("The hour %s, the %s of the %s, %s–%s", of(h.Ruler), chart.Ordinal(h.Number), part, clock(h.Start), clock(h.End)),
			fmt.Sprintf("Sunrise %s  ·  sunset %s  ·  next sunrise %s", clock(d.Sunrise), clock(d.Sunset), clock(d.NextSunrise)))
	} else {
		lines = append(lines, fmt.Sprintf("The day %s  ·  no planetary hour: %s", of(d.Ruler), d.Unavailable))
	}
	waxing := "waning"
	if s.Moon.Phase.Waxing() {
		waxing = "waxing"
	}
	lines = append(lines, fmt.Sprintf("%s Moon, %s, %.0f%% lit", s.Moon.Phase, waxing, s.Moon.Illuminated*100))
	return lines
}

// Table layout, in mm.
const (
	rowHeight = 6.0
	tableLeft = margin + 6
)

// tableRule draws a rule across the tables' width at y.
func (p *page) tableRule(y, x0, x1 float64) {
	p.ctx.SetFillColor(canvas.Transparent)
	p.ctx.SetStrokeColor(rule)
	p.ctx.SetStrokeWidth(0.25)
	p.ctx.MoveTo(x0, y)
	p.ctx.LineTo(x1, y)
	p.ctx.Stroke()
}

// tableHeader draws column headings at the given x positions with their
// baseline at y, and a rule beneath; it returns the y of the first row.
func (rep *report) tableHeader(p *page, y, x1 float64, cols []float64, names ...string) float64 {
	for i, n := range names {
		p.text(rep.f.bold, 9, muted, cols[i], y, n, canvas.Left)
	}
	p.tableRule(y-2, cols[0], x1)
	return y - 2 - rowHeight + 1.5
}

// bodies draws the table of planets and nodes from y down, and returns the
// y below it.
func (rep *report) bodies(p *page, s *sky.Sky, y float64) float64 {
	f := rep.f
	right := PageWidth - tableLeft
	cols := []float64{tableLeft, tableLeft + 8, tableLeft + 40, tableLeft + 80, tableLeft + 104}
	y = rep.tableHeader(p, y, right, cols, "", "Body", "Position", "Motion", "Condition")
	for _, b := range s.Bodies {
		rep.planetGlyph(p, b.Planet, cols[0]+2.5, y+1.2)
		p.text(f.body, 10.5, ink, cols[1], y, b.Planet.String(), canvas.Left)
		p.text(f.body, 10.5, ink, cols[2], y, b.Position(), canvas.Left)
		motion, col := "", ink
		switch {
		case b.Planet == zodiac.NorthNode || b.Planet == zodiac.SouthNode:
		case b.Retrograde:
			motion, col = "Retrograde", accent
		default:
			motion = "Direct"
		}
		p.text(f.italic, 10.5, col, cols[3], y, motion, canvas.Left)
		p.text(f.body, 10.5, ink, cols[4], y, condition(b), canvas.Left)
		y -= rowHeight
	}
	p.tableRule(y+rowHeight-2, cols[0], right)
	return y
}

// condition writes a body's dignities and closeness to the Sun.
func condition(b sky.Body) string {
	var parts []string
	for _, d := range b.Dignities {
		parts = append(parts, d.String())
	}
	if b.Solar != sky.NotNearSun {
		parts = append(parts, b.Solar.String())
	}
	return strings.Join(parts, ", ")
}

// planetGlyph draws a planet's symbol centred on x, y: from the symbol font,
// but the Sun's, which it lacks, as a circle with a point at its centre.
func (rep *report) planetGlyph(p *page, pl zodiac.Planet, x, y float64) {
	if pl == zodiac.Sun {
		p.ctx.SetFillColor(canvas.Transparent)
		p.ctx.SetStrokeColor(ink)
		p.ctx.SetStrokeWidth(0.3)
		p.ctx.DrawPath(x, y, canvas.Circle(1.5))
		p.ctx.SetFillColor(ink)
		p.ctx.SetStrokeColor(canvas.Transparent)
		p.ctx.DrawPath(x, y, canvas.Circle(0.35))
		return
	}
	p.text(rep.f.symbols, 10, ink, x, y-1.4, pl.Symbol(), canvas.Center)
}

// skyAspects draws the aspects in two columns from y down, closest first
// down the left then the right, and returns the y below them.
func (rep *report) skyAspects(p *page, s *sky.Sky, y float64) float64 {
	f := rep.f
	right := PageWidth - tableLeft
	if len(s.Aspects) == 0 {
		p.text(f.italic, 10.5, muted, PageWidth/2, y, "No aspects between the planets.", canvas.Center)
		return y - rowHeight
	}
	half := (right - tableLeft) / 2
	cols := func(x float64) []float64 { return []float64{x, x + 42, x + 55} }
	top := y
	for side := range 2 {
		c := cols(tableLeft + float64(side)*half)
		y = rep.tableHeader(p, top, c[0]+half-6, c, "Aspect", "Orb", "")
	}
	rows := (len(s.Aspects) + 1) / 2
	for i, a := range s.Aspects {
		c := cols(tableLeft + float64(i/rows)*half)
		ry := y - float64(i%rows)*rowHeight
		p.text(f.body, 10.5, ink, c[0], ry, a.A.String(), canvas.Left)
		x := c[0] + f.body.Face(10.5).TextWidth(a.A.String()) + 4
		rep.aspectGlyph(p, a.Kind, x, ry+1.2)
		p.text(f.body, 10.5, ink, x+4, ry, a.B.String(), canvas.Left)
		p.text(f.body, 10.5, ink, c[1], ry, dms(a.Orb), canvas.Left)
		motion := "separating"
		if a.Applying {
			motion = "applying"
		}
		p.text(f.italic, 10.5, muted, c[2], ry, motion, canvas.Left)
	}
	y -= float64(rows) * rowHeight
	p.tableRule(y+rowHeight-2, tableLeft, right)
	return y
}

// aspectGlyph draws a planetary aspect's symbol centred on x, y: the
// Square's vector symbols, and for the conjunction the symbol font's.
func (rep *report) aspectGlyph(p *page, k sky.Kind, x, y float64) {
	geomantic := map[sky.Kind]aspect.Aspect{
		sky.Opposition: aspect.Opposition, sky.Square: aspect.Square, sky.Trine: aspect.Trine, sky.Sextile: aspect.Sextile,
	}
	if a, ok := geomantic[k]; ok {
		p.glyph(a, x, y, 2.6)
		return
	}
	p.text(rep.f.symbols, 9, ink, x, y-1.3, k.Symbol(), canvas.Center)
}

// dms writes degrees as degrees and minutes: "2°14′".
func dms(deg float64) string {
	d := int(deg)
	return fmt.Sprintf("%d°%02d′", d, int((deg-float64(d))*60))
}

// lunations draws the New and Full Moons either side of the moment from y
// down, and returns the y below them.
func (rep *report) lunations(p *page, m sky.Moon, y float64) float64 {
	f := rep.f
	right := PageWidth - tableLeft
	cols := []float64{tableLeft, tableLeft + 24, tableLeft + 88}
	y = rep.tableHeader(p, y, right, cols, "Lunations", "New Moon", "Full Moon")
	when := func(t time.Time) string { return t.Format("2 January 2006, 3:04 pm MST") }
	for _, r := range []struct {
		name       string
		newM, full time.Time
	}{{"Previous", m.PreviousNew, m.PreviousFull}, {"Next", m.NextNew, m.NextFull}} {
		p.text(f.italic, 10.5, muted, cols[0], y, r.name, canvas.Left)
		p.text(f.body, 10.5, ink, cols[1], y, when(r.newM), canvas.Left)
		p.text(f.body, 10.5, ink, cols[2], y, when(r.full), canvas.Left)
		y -= rowHeight
	}
	p.tableRule(y+rowHeight-2, cols[0], right)
	return y
}
