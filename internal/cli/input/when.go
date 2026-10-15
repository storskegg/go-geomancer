package input

import (
	"errors"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/storskegg/go-geomancer/place"
	"github.com/storskegg/go-geomancer/when"
)

// Occasion collects when and where a reading took place: the local date
// and time and the querent's place, from flags or prompts, and the instants
// the local time can mean there.
type Occasion struct {
	DateText, TimeText, PlaceText string // as typed

	Date  when.Date
	Clock when.Clock
	Place place.Match

	// Candidates are the instants the local time means at the place:
	// usually one, two when clocks went back and the time happened twice.
	// Choice indexes the one chosen.
	Candidates []time.Time
	Choice     int
}

// Time returns the chosen instant, or the zero time if it isn't resolved.
func (o *Occasion) Time() time.Time {
	if o.Choice < 0 || o.Choice >= len(o.Candidates) {
		return time.Time{}
	}
	return o.Candidates[o.Choice]
}

// Resolve works out the instants the date and time mean at the place: its
// time zone, then when.Resolve. A time in the gap when clocks went forward
// is an error.
func (o *Occasion) Resolve() error {
	loc, err := place.TimeZone(o.Place.Place.Latitude, o.Place.Place.Longitude)
	if err != nil {
		return err
	}
	ts, err := when.Resolve(o.Date, o.Clock, loc)
	if err != nil {
		return err
	}
	o.Candidates, o.Choice = ts, 0
	return nil
}

// recheck resolves the time again if the place is already known, as when
// the querent goes back from the place to change the date or time.
func (o *Occasion) recheck() error {
	if o.Place.Label == "" {
		return nil
	}
	if err := o.Resolve(); err != nil {
		var gap *when.GapError
		if errors.As(err, &gap) {
			return message(fmt.Sprintf("At %s, %s.", o.Place.Label, err))
		}
		return err
	}
	return nil
}

// Ambiguous reports whether the local time happened twice, so the querent
// must say which.
func (o *Occasion) Ambiguous() bool { return len(o.Candidates) > 1 }

// message is an error shown to the querent under a prompt: a sentence,
// unlike Go's usual error strings.
type message string

func (m message) Error() string { return string(m) }

// liveInput is an input that describes its value as it's typed, such as
// the date it reads, below the input. It updates the description itself
// rather than through huh's DescriptionFunc, which evaluates on another
// goroutine.
type liveInput struct {
	*huh.Input
	value    *string
	describe func(string) string
}

func (l *liveInput) Update(msg tea.Msg) (huh.Model, tea.Cmd) {
	_, cmd := l.Input.Update(msg)
	l.Description(l.describe(*l.value))
	return l, cmd
}

// dateGroup asks for the date of the reading, writing it to o.
func dateGroup(o *Occasion, now func() time.Time, accessible bool) *huh.Group {
	const examples = "Such as 15 October 2026, 10/15/2026 or today."
	in := huh.NewInput().
		Title("Date").
		Description(examples).
		Value(&o.DateText).
		Validate(func(s string) error {
			d, err := when.ParseDate(s, now())
			if err != nil {
				return message("Not a date I can read. " + examples)
			}
			o.Date = d
			return o.recheck()
		})
	if accessible {
		return huh.NewGroup(in)
	}
	return huh.NewGroup(&liveInput{Input: in, value: &o.DateText, describe: func(s string) string {
		if strings.TrimSpace(s) == "" {
			return examples
		}
		if d, err := when.ParseDate(s, now()); err == nil {
			return "→ " + d.String()
		}
		return "→ …"
	}})
}

// timeGroup asks for the local time of the reading, writing it to o.
func timeGroup(o *Occasion, accessible bool) *huh.Group {
	const examples = "The local time, such as 3:30 pm or 15:30."
	in := huh.NewInput().
		Title("Time").
		Description(examples).
		Value(&o.TimeText).
		Validate(func(s string) error {
			c, err := when.ParseClock(s)
			if err != nil {
				return message("Not a time I can read. " + examples)
			}
			o.Clock = c
			return o.recheck()
		})
	if accessible {
		return huh.NewGroup(in)
	}
	return huh.NewGroup(&liveInput{Input: in, value: &o.TimeText, describe: func(s string) string {
		if strings.TrimSpace(s) == "" {
			return examples
		}
		if c, err := when.ParseClock(s); err == nil {
			return "→ " + c.String()
		}
		return "→ …"
	}})
}

// placeMatches is how many matches the place field lists.
const placeMatches = 8

// placeGroup asks for the querent's place, writing it to o and resolving
// the date and time there. It validates against the time: a time that
// didn't occur at the place, when clocks went forward, is an error, so the
// querent can go back and change it.
func placeGroup(o *Occasion, places func() *place.Index, accessible bool) *huh.Group {
	in := huh.NewInput().
		Title("Place").
		Description("Where the querent was: a city or town, such as Springfield, IL.").
		Value(&o.PlaceText).
		Validate(func(s string) error {
			ix := places()
			m, ok := ix.Find(s)
			if !ok {
				// Accessible prompts have no list to choose from, so take
				// the best match.
				ms := ix.Search(s, 1)
				if len(ms) == 0 {
					return message("No place by that name. Type a city or town and its state, such as Springfield, IL.")
				}
				m = ms[0]
			}
			o.Place = m
			if err := o.Resolve(); err != nil {
				var gap *when.GapError
				if errors.As(err, &gap) {
					return message(fmt.Sprintf("At %s, %s. Go back to change the time.", m.Label, err))
				}
				return err
			}
			return nil
		})
	if accessible {
		return huh.NewGroup(in)
	}
	return huh.NewGroup(&placeField{Input: in, places: places, query: &o.PlaceText})
}

// placeField is an input that lists the places matching what's typed, most
// populous first: ↑ and ↓ move through them, and Enter takes the one
// highlighted. huh's Select draws every option on every frame, too slow for
// 168,000 places, so this searches as it goes and shows only the best.
type placeField struct {
	*huh.Input
	places   func() *place.Index
	query    *string
	matches  []place.Match
	cursor   int
	searched string // the query matches are for
	// goingBack is set while the querent goes back to an earlier prompt, as
	// the time error tells them to: huh won't leave a group with an error.
	goingBack bool
}

func (p *placeField) Focus() tea.Cmd {
	p.goingBack = false
	return p.Input.Focus()
}

func (p *placeField) Error() error {
	if p.goingBack {
		return nil
	}
	return p.Input.Error()
}

var (
	matchStyle    = lipgloss.NewStyle().PaddingLeft(2)
	selectedStyle = lipgloss.NewStyle().Bold(true)
)

func (p *placeField) Update(msg tea.Msg) (huh.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "shift+tab":
			p.goingBack = true
		case "up":
			p.cursor = max(p.cursor-1, 0)
			return p, nil
		case "down":
			p.cursor = min(p.cursor+1, len(p.matches)-1)
			return p, nil
		case "enter", "tab":
			if p.cursor >= 0 && p.cursor < len(p.matches) && *p.query != p.matches[p.cursor].Label {
				*p.query = p.matches[p.cursor].Label
				p.Value(p.query) // show the label before validating it
				p.searched = *p.query
			}
		}
	}
	_, cmd := p.Input.Update(msg)
	if *p.query != p.searched {
		p.searched = *p.query
		p.matches, p.cursor = p.places().Search(*p.query, placeMatches), 0
	}
	return p, cmd
}

// View draws the input with the matches below it, always placeMatches
// lines of them: huh sizes a group's view from its fields' first views, so
// the list must take its full height from the start.
func (p *placeField) View() string {
	lines := []string{p.Input.View()}
	for i, m := range p.matches {
		if i == p.cursor {
			lines = append(lines, matchStyle.Render(selectedStyle.Render("> "+m.Label)))
		} else {
			lines = append(lines, matchStyle.Render("  "+m.Label))
		}
	}
	if strings.TrimSpace(*p.query) != "" && len(p.matches) == 0 {
		lines = append(lines, matchStyle.Render("  No matches"))
	}
	for len(lines) < 1+placeMatches {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

// zoneLabel describes one meaning of a time that happened twice.
func zoneLabel(t time.Time, first bool) string {
	which := "the second time, after the clocks went back"
	if first {
		which = "the first time, before the clocks went back"
	}
	return fmt.Sprintf("%s (%s)", t.Format("3:04 pm MST"), which)
}

// zoneGroup asks which of two meanings a local time has, when clocks went
// back and it happened twice. It's shown only then.
func zoneGroup(o *Occasion) *huh.Group {
	// Without a height a select shows one option: this fits the title, the
	// description and both times.
	sel := huh.NewSelect[int]().
		Title("Which time?").
		Description("The clocks went back that night, so this time happened twice.").
		Value(&o.Choice).
		Height(2 + 2)
	return huh.NewGroup(&zoneSelect{Select: sel, o: o}).WithHideFunc(func() bool { return !o.Ambiguous() })
}

// zoneSelect fills its options when it gains focus, once the place, and so
// the candidate times, are known.
type zoneSelect struct {
	*huh.Select[int]
	o *Occasion
}

func (z *zoneSelect) Focus() tea.Cmd {
	opts := make([]huh.Option[int], len(z.o.Candidates))
	for i, t := range z.o.Candidates {
		opts[i] = huh.NewOption(zoneLabel(t, i == 0), i)
	}
	z.Options(opts...)
	return z.Select.Focus()
}

func (z *zoneSelect) Update(msg tea.Msg) (huh.Model, tea.Cmd) {
	_, cmd := z.Select.Update(msg)
	return z, cmd
}

// accessibleZoneGroup asks which time, for accessible prompts, which show
// every group: it's built only when the time is ambiguous.
func accessibleZoneGroup(o *Occasion) *huh.Group {
	opts := make([]huh.Option[int], len(o.Candidates))
	for i, t := range o.Candidates {
		opts[i] = huh.NewOption(zoneLabel(t, i == 0), i)
	}
	return huh.NewGroup(huh.NewSelect[int]().
		Title("Which time? The clocks went back that night, so this time happened twice.").
		Options(opts...).
		Value(&o.Choice))
}
