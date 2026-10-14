package report

import (
	"fmt"
	"strings"

	"github.com/tdewolff/canvas"

	"github.com/storskegg/go-geomancer/aspect"
)

// reading draws the Reading from y down: for each house bearing on the
// question (see chart.Reading.Focus), its parts and the meaning of its
// figure. It continues onto new pages as needed.
func (rep *report) reading(p *page, y float64) {
	const gap = 5.0 // between entries
	width := PageWidth - 2*margin
	f := rep.f

	// Start on a new page rather than leave the heading alone at the foot.
	if y < bottom+40 {
		p = rep.newPage()
		y = PageHeight - 34
	}
	y = rep.heading(p, y, "The Reading") - 5

	for _, fh := range rep.r.Focus() {
		parts := []string{fmt.Sprintf("House %d", fh.House)}
		if fh.Querent {
			parts = append(parts, "Querent")
		}
		if fh.Significator {
			parts = append(parts, "Significator")
		}
		if fh.Aspect != nil {
			s := fh.Aspect.Aspect.String()
			if fh.Aspect.Side != aspect.Neither {
				s += " " + strings.ToLower(fh.Aspect.Side.String())
			}
			parts = append(parts, s)
		}
		if sig, ok := rep.opts.Content.Signification(fh.House); ok {
			parts = append(parts, sig.Title)
		}
		col := ink
		if fh.Significator {
			col = accent
		}
		head := canvas.NewRichText(f.bold.Face(10, col))
		head.WriteString(strings.Join(parts, "  ·  "))

		fig := rep.r.Square[fh.House-1].Figure
		meaning := rep.opts.Content.For(fig).House(fh.House)
		if meaning == "" {
			meaning = "Not yet written."
		}
		body := canvas.NewRichText(f.body.Face(10.5, ink))
		body.WriteFace(f.italic.Face(10.5, ink), fig.Name())
		body.WriteString(" — " + meaning)

		headText, bodyText := wrap(head, width, canvas.Left), wrap(body, width, canvas.Left)
		need := headText.Bounds().H() + 1.5 + bodyText.Bounds().H()
		if y-need < bottom {
			p = rep.newPage()
			y = PageHeight - 34
		}
		y = p.paragraph(headText, margin, y) - 1.5
		y = p.paragraph(bodyText, margin, y) - gap
	}
}
