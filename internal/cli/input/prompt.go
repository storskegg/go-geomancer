package input

import (
	"errors"
	"fmt"
	"io"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/internal/cli/render"
)

// ErrAborted is returned when the user cancels a prompt.
var ErrAborted = errors.New("cancelled")

// Prompt asks for each Mother in turn with a scrollable select. Typing
// filters the list by name or dot pattern.
type Prompt struct {
	In  io.Reader
	Out io.Writer
	// Accessible replaces the interactive select with numbered, line-based
	// prompts suitable for screen readers.
	Accessible bool
}

var ordinals = [4]string{"first", "second", "third", "fourth"}

// selectHeight fits the filter line, the preview and ten figures.
const selectHeight = 1 + 5 + 10

func (p Prompt) Mothers() ([4]figure.Figure, error) {
	var m [4]figure.Figure
	in := p.In
	if p.Accessible {
		in = byteReader{in}
	}
	form := newForm(&m, p.Accessible).WithInput(in).WithOutput(p.Out)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return m, ErrAborted
		}
		return m, err
	}
	return m, nil
}

// newForm builds a form with one select per Mother, writing choices to m.
func newForm(m *[4]figure.Figure, accessible bool) *huh.Form {
	all := figure.All()
	options := make([]huh.Option[figure.Figure], 0, len(all))
	for _, f := range all {
		options = append(options, huh.NewOption(fmt.Sprintf("%-14s  %s", f.Name(), f.Pattern()), f))
	}

	groups := make([]*huh.Group, 0, len(m))
	for i := range m {
		m[i] = all[0]
		title := fmt.Sprintf("Mother %d (M%d)", i+1, i+1)
		sel := huh.NewSelect[figure.Figure]().
			Options(options...).
			Value(&m[i])
		if accessible {
			sel.Title(title).Description(fmt.Sprintf("Choose the %s Mother.", ordinals[i]))
			groups = append(groups, huh.NewGroup(sel))
			continue
		}
		// While filtering, the select shows its filter input in place of its
		// title, so the Mother's name goes in a note above it. The value tracks the
		// cursor, so the preview follows it.
		sel.Title("Figure").
			Filtering(true).
			DescriptionFunc(func() string { return preview(m[i]) }, &m[i]).
			Height(selectHeight)
		note := huh.NewNote().
			Title(title).
			Description(fmt.Sprintf("Choose the %s Mother. Type to filter by name or pattern.", ordinals[i]))
		groups = append(groups, huh.NewGroup(note, followSelect{sel, &m[i]}))
	}

	return huh.NewForm(groups...).
		WithTheme(huh.ThemeFunc(huh.ThemeBase16)).
		WithAccessible(accessible)
}

// followSelect keeps the bound value on the hovered option. huh's Select
// only updates it when the cursor moves, not when filtering moves the cursor,
// which would leave the preview showing a figure that's been filtered out.
type followSelect struct {
	*huh.Select[figure.Figure]
	value *figure.Figure
}

func (s followSelect) Update(msg tea.Msg) (huh.Model, tea.Cmd) {
	_, cmd := s.Select.Update(msg)
	if f, ok := s.Hovered(); ok {
		*s.value = f
	}
	return s, cmd
}

// byteReader reads at most one byte at a time. Accessible prompts each wrap
// the input in their own buffered scanner, so without this the first prompt
// would consume every line of piped input and the rest would silently take
// their defaults.
type byteReader struct{ r io.Reader }

func (b byteReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		p = p[:1]
	}
	return b.r.Read(p)
}

// preview shows a figure's dots beside its name and planet.
func preview(f figure.Figure) string {
	glyph := strings.Split(render.Glyph(f), "\n")
	info := []string{f.Name(), f.Pattern(), f.Planet().String(), ""}
	lines := make([]string, len(glyph))
	for i := range glyph {
		lines[i] = glyph[i] + "   " + info[i]
	}
	return strings.Join(lines, "\n")
}
