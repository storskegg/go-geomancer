package input

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/internal/cli/render"
	"github.com/storskegg/go-geomancer/position"
	"github.com/storskegg/go-geomancer/system"
)

// ErrAborted is returned when the user cancels a prompt.
var ErrAborted = errors.New("cancelled")

// Prompt asks for each Mother in turn with a scrollable select. Typing
// filters the list by name or dot pattern. It can first ask for the system
// and the significator.
type Prompt struct {
	In  io.Reader
	Out io.Writer
	// Accessible replaces the interactive select with numbered, line-based
	// prompts suitable for screen readers.
	Accessible bool
	// System is the system the Mothers are cast under; previews show its
	// correspondences. If Systems is non-empty, the user first chooses one
	// of them, starting from *System, and the choice is written to System.
	System  *system.System
	Systems []system.System
	// If Significations is non-empty, the user chooses the significator,
	// the house of the question, from them; the choice (1–12) is written to
	// Significator, whose value, if valid, is the initial choice.
	Significator   *int
	Significations []content.Signification
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
	if len(p.Significations) > 0 {
		groups = append(groups, significatorGroup(p.Significator, p.Significations, p.Accessible))
	}
	groups = append(groups, motherGroups(&m, p.System, p.Accessible)...)
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

// significatorGroup offers the houses' significations, writing the chosen
// house to n. Typing filters by title and signification, not by house
// number: the list shows numbers as circled digits (⑦), which typed digits
// don't match. Accessible prompts number the list 1–12 themselves.
func significatorGroup(n *int, sigs []content.Signification, accessible bool) *huh.Group {
	options := make([]huh.Option[int], 0, len(sigs))
	for _, s := range sigs {
		label := s.Title + " — " + s.Text
		if !accessible {
			label = circled(s.House) + " " + label
		}
		options = append(options, huh.NewOption(label, s.House))
	}
	if *n < 1 || *n > len(sigs) {
		*n = sigs[0].House
	}
	sel := huh.NewSelect[int]().Options(options...).Value(n)
	if accessible {
		sel.Title("Significator").Description("Choose the house of what you're asking about.")
		return huh.NewGroup(sel)
	}
	// As for the Mothers, the filter takes the select's title slot, so the
	// prompt's own title goes in a note.
	sel.Title("House").Filtering(true).Height(1 + len(sigs))
	note := huh.NewNote().
		Title("Significator").
		Description("Choose the house of what you're asking about. Type to filter by subject.")
	return huh.NewGroup(note, sel)
}

// circled returns n (1–20) as a circled digit, ① to ⑳.
func circled(n int) string { return string(rune('①' + n - 1)) }

// motherGroups builds one select per Mother, writing choices to m. sys,
// which may be nil, supplies the previews' planets.
func motherGroups(m *[4]figure.Figure, sys *system.System, accessible bool) []*huh.Group {
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
		// cursor, so the preview follows it, and the system, which the user may
		// go back and change.
		sel.Title("Figure").
			Filtering(true).
			DescriptionFunc(func() string { return preview(m[i], deref(sys)) }, []any{&m[i], sys}).
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

func deref(sys *system.System) system.System {
	if sys == nil {
		return nil
	}
	return *sys
}

// preview shows a figure's dots beside its name, pattern and, if sys is
// non-nil and gives it one, planet.
func preview(f figure.Figure, sys system.System) string {
	planet := ""
	if sys != nil {
		if p, ok := sys.Planet(f); ok {
			planet = p.String()
		}
	}
	glyph := strings.Split(render.Glyph(f), "\n")
	info := []string{f.Name(), f.Pattern(), planet, ""}
	lines := make([]string, len(glyph))
	for i := range glyph {
		lines[i] = glyph[i] + "   " + info[i]
	}
	return strings.Join(lines, "\n")
}
