package report

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/system/goldendawn"
	"github.com/storskegg/go-geomancer/zodiac"
)

// TestPreview renders sample reports as PNGs and PDFs into the directory in
// REPORT_PREVIEW_DIR, for checking the design by eye. It's skipped
// otherwise.
func TestPreview(t *testing.T) {
	dir := os.Getenv("REPORT_PREVIEW_DIR")
	if dir == "" {
		t.Skip("set REPORT_PREVIEW_DIR to render previews")
	}
	chicago, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatal(err)
	}
	circumstances := &chart.Circumstances{
		Time: time.Date(2026, 10, 15, 15, 30, 0, 0, chicago), TimeZone: "America/Chicago", Zone: "CDT",
		Place: chart.Location{Name: "Springfield", State: "IL", Latitude: 39.80172, Longitude: -89.64371},
	}
	for name, sig := range map[string]int{"significator": 10, "plain": 0} {
		r, err := chart.NewReading([4]figure.Figure{figure.Amissio, figure.FortunaMinor, figure.Via, figure.Populus}, goldendawn.System)
		if err != nil {
			t.Fatal(err)
		}
		if err := r.SetSignificator(sig); err != nil {
			t.Fatal(err)
		}
		if sig != 0 {
			r.Question = "Will I be offered the position I interviewed for last week, and will it be a good move for me?"
			r.Circumstances = circumstances
			if r.Sky, err = circumstances.Sky(); err != nil {
				t.Fatal(err)
			}
			r.Timing = chart.Timing(r.Sky, "career", []zodiac.Planet{zodiac.Sun})
		}
		pages, err := Pages(r, Options{})
		if err != nil {
			t.Fatal(err)
		}
		for i, p := range pages {
			writePNG(t, filepath.Join(dir, fmt.Sprintf("%s-%d.png", name, i+1)), p)
		}
		f, err := os.Create(filepath.Join(dir, name+".pdf"))
		if err != nil {
			t.Fatal(err)
		}
		if err := WritePDF(f, r, Options{}); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func writePNG(t *testing.T, path string, c *canvas.Canvas) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if err := png.Encode(f, rasterizer.Draw(c, canvas.DPMM(5), canvas.DefaultColorSpace)); err != nil {
		t.Fatal(err)
	}
}
