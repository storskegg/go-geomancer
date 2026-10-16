package render

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/sky"
	"github.com/storskegg/go-geomancer/zodiac"
)

// Heavens describes the astrology of the moment: the planetary day and hour
// and the Moon's phase, then tables of the planets and nodes, their aspects,
// and the lunations either side.
func Heavens(s *sky.Sky, opts Options) string {
	return strings.Join([]string{
		opts.heavensSummary(s),
		opts.bodiesTable(s),
		opts.skyAspectsTable(s),
		lunationsTable(s.Moon),
	}, "\n\n")
}

// clock writes a time of day: "7:11 am".
func clock(t time.Time) string { return t.Format("3:04 pm") }

func (o Options) heavensSummary(s *sky.Sky) string {
	var lines []string
	line := func(label, text string) {
		lines = append(lines, labelStyle.Render(fmt.Sprintf("%-9s ", label))+text)
	}
	d := s.Day
	day := o.planetName(d.Ruler)
	if !d.Sunrise.IsZero() {
		day += mutedStyle.Render(fmt.Sprintf(" · %s, from sunrise at %s", d.Sunrise.Format("Monday"), clock(d.Sunrise)))
	} else {
		day += mutedStyle.Render(" · " + s.Time.Format("Monday"))
	}
	line("Day", day)
	if h := d.Hour; h != nil {
		part := "day"
		if h.Night {
			part = "night"
		}
		line("Hour", o.planetName(h.Ruler)+mutedStyle.Render(fmt.Sprintf(" · %s hour of the %s, %s–%s", chart.Ordinal(h.Number), part, clock(h.Start), clock(h.End))))
	} else {
		line("Hour", mutedStyle.Render("not available: "+d.Unavailable))
	}
	waxing := "waning"
	if s.Moon.Phase.Waxing() {
		waxing = "waxing"
	}
	line("Moon", fmt.Sprintf("%s · %s · %.0f%% lit", s.Moon.Phase, waxing, s.Moon.Illuminated*100))
	return strings.Join(lines, "\n")
}

// planetName names a planet with its symbol: "☉ Sun", or with ASCII symbols
// just "Sun".
func (o Options) planetName(p zodiac.Planet) string {
	if o.Symbols == ASCII {
		return p.String()
	}
	return p.Symbol() + " " + p.String()
}

func (o Options) bodiesTable(s *sky.Sky) string {
	t := newTable("Body", "Position", "Motion", "Condition")
	for _, b := range s.Bodies {
		motion := mutedStyle.Render(none)
		switch {
		case b.Planet == zodiac.NorthNode || b.Planet == zodiac.SouthNode:
		case b.Retrograde:
			motion = "Retrograde"
		default:
			motion = "Direct"
		}
		t.Row(o.planetName(b.Planet), b.Position(), motion, condition(b))
	}
	return t.Render()
}

// condition writes a body's dignities and closeness to the Sun:
// "Detriment · Under the beams". Peregrine is muted.
func condition(b sky.Body) string {
	var parts []string
	for _, d := range b.Dignities {
		if d == sky.Peregrine {
			parts = append(parts, mutedStyle.Render(d.String()))
			continue
		}
		parts = append(parts, d.String())
	}
	if b.Solar != sky.NotNearSun {
		parts = append(parts, b.Solar.String())
	}
	if len(parts) == 0 {
		return mutedStyle.Render(none)
	}
	return strings.Join(parts, mutedStyle.Render(" · "))
}

func (o Options) skyAspectsTable(s *sky.Sky) string {
	if len(s.Aspects) == 0 {
		return mutedStyle.Render("No aspects between the planets.")
	}
	t := newTable("Aspect", "Orb", "")
	for _, a := range s.Aspects {
		kind := a.Kind.Symbol()
		if o.Symbols == ASCII {
			kind = strings.ToLower(a.Kind.String())
		}
		var style lipgloss.Style
		switch a.Kind {
		case sky.Conjunction:
			style = nameStyle
		case sky.Square, sky.Opposition:
			style = stressfulStyle
		case sky.Sextile, sky.Trine:
			style = easingStyle
		}
		motion := "separating"
		if a.Applying {
			motion = "applying"
		}
		t.Row(fmt.Sprintf("%s %s %s", a.A, style.Render(kind), a.B), dms(a.Orb), mutedStyle.Render(motion))
	}
	return t.Render()
}

// dms writes degrees as degrees and minutes: "2°14′".
func dms(deg float64) string {
	d := int(deg)
	return fmt.Sprintf("%d°%02d′", d, int((deg-float64(d))*60))
}

func lunationsTable(m sky.Moon) string {
	when := func(t time.Time) string { return t.Format("Mon 2 Jan 2006, 3:04 pm MST") }
	t := newTable("", "New Moon", "Full Moon")
	t.Row("Previous", when(m.PreviousNew), when(m.PreviousFull))
	t.Row("Next", when(m.NextNew), when(m.NextFull))
	return t.Render()
}
