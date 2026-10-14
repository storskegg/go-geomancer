package input

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/internal/cli/render"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/system"
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
	// If Systems is non-empty, the user first chooses one of them. System
	// holds the initial choice and receives the final one.
	Systems []system.System
	System  *system.System
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
	var groups []*huh.Group
	if len(p.Systems) > 0 {
		groups = append(groups, systemGroup(p.System, p.Systems, p.Accessible))
	}
	groups = append(groups, motherGroups(&m, p.Accessible)...)
	form := newForm(groups...).WithAccessible(p.Accessible).WithInput(in).WithOutput(p.Out)
	return m, formError(form.Run())
}

// formError translates huh's cancellation error to ErrAborted.
func formError(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		return ErrAborted
	}
	return err
}

func newForm(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).WithTheme(huh.ThemeFunc(huh.ThemeBase16))
}

// systemGroup offers a choice of systems, writing the choice to sys.
func systemGroup(sys *system.System, systems []system.System, accessible bool) *huh.Group {
	options := make([]huh.Option[system.System], 0, len(systems))
	for _, s := range systems {
		options = append(options, huh.NewOption(s.Name(), s))
	}
	if *sys == nil {
		*sys = systems[0]
	}
	// Without a height the select shows a single option. This fits the
	// title, the two-line description and every system.
	sel := huh.NewSelect[system.System]().
		Title("System").
		Options(options...).
		Value(sys).
		Height(1 + 2 + len(systems))
	if accessible {
		sel.Description("Choose the system of correspondences.")
	} else {
		sel.DescriptionFunc(func() string { return describeSystem(*sys) }, sys)
	}
	return huh.NewGroup(sel)
}

// describeSystem summarises how a system differs: its segment elements and
// where it places the Mothers.
func describeSystem(s system.System) string {
	segs := make([]string, 0, 4)
	for _, seg := range figure.Segments() {
		segs = append(segs, fmt.Sprintf("%v %v", seg, s.SegmentElement(seg)))
	}
	houses := make([]string, 0, 4)
	for _, p := range []position.Position{position.M1, position.M2, position.M3, position.M4} {
		h, _ := s.House(p)
		houses = append(houses, strconv.Itoa(h))
	}
	return strings.Join(segs, " · ") + "\nMothers in houses " + strings.Join(houses, ", ")
}

// motherGroups builds one select per Mother, writing choices to m.
func motherGroups(m *[4]figure.Figure, accessible bool) []*huh.Group {
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
	return groups
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
