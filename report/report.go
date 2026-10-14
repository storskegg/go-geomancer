// Package report draws a reading as a printed report: US Letter pages with
// the Astrological Square as a traditional square chart, the Shield, and the
// Reading, in a classical style with embedded fonts.
package report

import (
	"fmt"
	"io"
	"time"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/pdf"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/system"
)

// US Letter, in millimetres.
const (
	PageWidth  = 215.9
	PageHeight = 279.4
)

// bottom is the lowest a page's text may reach, above the page number.
const bottom = 28.0

// Options adjust a report.
type Options struct {
	// Content supplies significations and meanings; nil means
	// content.Default().
	Content *content.Set
	// Time is when the chart was cast, shown in the header; the zero time
	// omits it.
	Time time.Time
	// Nephews calls N1–N4 Nephews rather than Nieces.
	Nephews bool
}

// report is the state of a report being drawn.
type report struct {
	r     chart.Reading
	sys   system.System
	opts  Options
	f     *fonts
	pages []*page
}

// Pages draws the report. The reading's system must be registered.
func Pages(r chart.Reading, opts Options) ([]*canvas.Canvas, error) {
	sys, ok := system.Get(r.System)
	if !ok {
		return nil, fmt.Errorf("report: unknown system %q", r.System)
	}
	f, err := loadFonts()
	if err != nil {
		return nil, err
	}
	if opts.Content == nil {
		opts.Content = content.Default()
	}
	rep := &report{r: r, sys: sys, opts: opts, f: f}

	p := rep.newPage()
	y := rep.header(p)
	y = rep.square(p, y-6)
	rep.squareLegend(p, y-7)

	p = rep.newPage()
	y = rep.heading(p, PageHeight-34, "The Shield")
	y = rep.shield(p, y-6)
	rep.reading(p, y-10)

	canvases := make([]*canvas.Canvas, len(rep.pages))
	for i, p := range rep.pages {
		canvases[i] = p.c
	}
	return canvases, nil
}

// WritePDF draws the report and writes it to w as a PDF.
func WritePDF(w io.Writer, r chart.Reading, opts Options) error {
	pages, err := Pages(r, opts)
	if err != nil {
		return err
	}
	out := pdf.New(w, PageWidth, PageHeight, nil)
	title := "Geomantic Reading"
	if r.Question != "" {
		title += ": " + r.Question
	}
	out.SetInfo(title, "", "geomancy", "", "geomancer")
	for i, p := range pages {
		if i > 0 {
			out.NewPage(PageWidth, PageHeight)
		}
		p.RenderTo(out)
	}
	return out.Close()
}

// page is one page being drawn.
type page struct {
	c   *canvas.Canvas
	ctx *canvas.Context
}

// newPage starts a page: parchment, a double-rule border and its number.
func (rep *report) newPage() *page {
	c := canvas.New(PageWidth, PageHeight)
	p := &page{c: c, ctx: canvas.NewContext(c)}
	rep.pages = append(rep.pages, p)

	ctx := p.ctx
	ctx.SetFillColor(parchment)
	ctx.SetStrokeColor(canvas.Transparent)
	ctx.DrawPath(0, 0, canvas.Rectangle(PageWidth, PageHeight))
	ctx.SetFillColor(canvas.Transparent)
	ctx.SetStrokeColor(rule)
	ctx.SetStrokeWidth(0.6)
	ctx.DrawPath(9, 9, canvas.Rectangle(PageWidth-18, PageHeight-18))
	ctx.SetStrokeWidth(0.25)
	ctx.DrawPath(11, 11, canvas.Rectangle(PageWidth-22, PageHeight-22))

	p.text(rep.f.italic, 9, muted, PageWidth/2, 16, fmt.Sprint(len(rep.pages)), canvas.Center)
	return p
}
