package report

import (
	"embed"
	"fmt"
	"image/color"
	"sync"

	"github.com/tdewolff/canvas"
)

//go:embed fonts/*.ttf
var fontFiles embed.FS

// fonts are the report's typefaces, all embedded.
type fonts struct {
	title   *canvas.FontFamily // Libre Caslon Display
	body    *canvas.FontFamily // Crimson Text
	italic  *canvas.FontFamily
	bold    *canvas.FontFamily // Crimson Text SemiBold
	symbols *canvas.FontFamily // Noto Sans Symbols: signs and elements
}

var loadFonts = sync.OnceValues(func() (*fonts, error) {
	load := func(file string) (*canvas.FontFamily, error) {
		b, err := fontFiles.ReadFile("fonts/" + file)
		if err != nil {
			return nil, err
		}
		f := canvas.NewFontFamily(file)
		if err := f.LoadFont(b, 0, canvas.FontRegular); err != nil {
			return nil, fmt.Errorf("report: font %s: %w", file, err)
		}
		return f, nil
	}
	var f fonts
	var err error
	for _, x := range []struct {
		dst  **canvas.FontFamily
		file string
	}{
		{&f.title, "LibreCaslonDisplay-Regular.ttf"},
		{&f.body, "CrimsonText-Regular.ttf"},
		{&f.italic, "CrimsonText-Italic.ttf"},
		{&f.bold, "CrimsonText-SemiBold.ttf"},
		{&f.symbols, "NotoSansSymbols.ttf"},
	} {
		if *x.dst, err = load(x.file); err != nil {
			return nil, err
		}
	}
	return &f, nil
})

func rgb(hex uint32) color.RGBA {
	return color.RGBA{uint8(hex >> 16), uint8(hex >> 8), uint8(hex), 0xff}
}

// The classical palette: parchment, dark-brown ink and earthy accents.
var (
	parchment = rgb(0xF4ECD8)
	ink       = rgb(0x2E2416)
	muted     = rgb(0x7A6A55)
	rule      = rgb(0x6B5536)
	sigShade  = rgb(0xE6D4AE) // the significator's house
	accent    = rgb(0x7A1F1F) // the significator, oxblood
	stressful = rgb(0x8E2C2C) // Opposition, Square
	easing    = rgb(0x4E6B3A) // Trine, Sextile
	mixed     = rgb(0xA8822C) // a mixed judgement of the timing, ochre

	// elementColours is indexed by zodiac.Element: Fire, Air, Water, Earth.
	elementColours = [4]color.RGBA{rgb(0xA33A2A), rgb(0xA8822C), rgb(0x2F5B7C), rgb(0x4E6B3A)}
)

const (
	lineWidth = 0.45 // mm, the Square's inner lines
	margin    = 24.0 // mm, left and right text margins
)
