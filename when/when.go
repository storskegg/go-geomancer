// Package when reads dates and times of day as people write them, and
// resolves a local date and time in a time zone to the instants it can
// mean: usually one, none in the gap when clocks go forward, and two in the
// overlap when they go back.
package when

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // time zones work without the system's database
)

// Date is a calendar date.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// Clock is a time of day, with an optional zone abbreviation ("CST") to
// choose between the two meanings of a time when clocks go back.
type Clock struct {
	Hour, Minute, Second int
	Zone                 string
}

var (
	// ErrDate and ErrClock are returned for text that isn't a date or a
	// time of day.
	ErrDate  = errors.New("not a date")
	ErrClock = errors.New("not a time")
)

func (d Date) String() string {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC).Format("Monday, 2 January 2006")
}

func (c Clock) String() string {
	t := time.Date(2000, 1, 1, c.Hour, c.Minute, c.Second, 0, time.UTC)
	layout := "3:04 pm"
	if c.Second != 0 {
		layout = "3:04:05 pm"
	}
	s := t.Format(layout)
	if c.Zone != "" {
		s += " " + c.Zone
	}
	return s
}

var months = map[string]time.Month{}

func init() {
	for m := time.January; m <= time.December; m++ {
		name := strings.ToLower(m.String())
		months[name] = m
		months[name[:3]] = m
	}
	months["sept"] = time.September
}

var weekdays = map[string]bool{}

func init() {
	for d := time.Sunday; d <= time.Saturday; d++ {
		name := strings.ToLower(d.String())
		weekdays[name], weekdays[name[:3]] = true, true
	}
	for _, w := range []string{"tues", "weds", "thur", "thurs"} {
		weekdays[w] = true
	}
}

// fields splits text into lowercase runs of letters and of digits, dropping
// everything else: "Oct. 15, 2026" gives "oct", "15", "2026".
func fields(s string) []string {
	var out []string
	cur := []rune{}
	kind := 0 // 1 letters, 2 digits
	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = cur[:0]
		}
	}
	for _, r := range strings.ToLower(s) {
		k := 0
		switch {
		case r >= 'a' && r <= 'z':
			k = 1
		case r >= '0' && r <= '9':
			k = 2
		}
		if k != kind {
			flush()
			kind = k
		}
		if k != 0 {
			cur = append(cur, r)
		}
	}
	flush()
	return out
}

// ParseDate reads a date: "2026-10-15", "10/15/2026", "10/15/26", "Oct 15
// 2026", "October 15, 2026", "15 October 2026", optionally with a weekday,
// or "today", "yesterday" or "tomorrow", relative to now. Numeric dates with
// slashes are month first. Two-digit years are 2000–2099.
func ParseDate(s string, now time.Time) (Date, error) {
	bad := fmt.Errorf("%w: %q", ErrDate, strings.TrimSpace(s))
	f := fields(s)
	// A leading weekday is allowed, and ignored.
	if len(f) > 0 && weekdays[f[0]] {
		f = f[1:]
	}
	if len(f) == 1 {
		switch f[0] {
		case "today":
			return dateOf(now), nil
		case "yesterday":
			return dateOf(now.AddDate(0, 0, -1)), nil
		case "tomorrow":
			return dateOf(now.AddDate(0, 0, 1)), nil
		}
	}
	if len(f) != 3 {
		return Date{}, bad
	}

	num := func(i int) (int, bool) {
		n, err := strconv.Atoi(f[i])
		return n, err == nil
	}
	var y, d int
	var m time.Month
	switch {
	case months[f[0]] != 0: // Oct 15 2026
		m = months[f[0]]
		dd, ok1 := num(1)
		yy, ok2 := num(2)
		if !ok1 || !ok2 {
			return Date{}, bad
		}
		d, y = dd, yy
	case months[f[1]] != 0: // 15 Oct 2026
		m = months[f[1]]
		dd, ok1 := num(0)
		yy, ok2 := num(2)
		if !ok1 || !ok2 {
			return Date{}, bad
		}
		d, y = dd, yy
	default:
		a, ok1 := num(0)
		b, ok2 := num(1)
		c, ok3 := num(2)
		if !ok1 || !ok2 || !ok3 {
			return Date{}, bad
		}
		if len(f[0]) == 4 { // 2026-10-15
			y, m, d = a, time.Month(b), c
		} else { // 10/15/2026
			m, d, y = time.Month(a), b, c
		}
	}
	if y < 100 && len(strconv.Itoa(y)) <= 2 {
		y += 2000
	}
	if m < time.January || m > time.December || d < 1 || y < 1 || y > 9999 {
		return Date{}, bad
	}
	// Reject days past the month's end, such as 31 April.
	if t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC); t.Day() != d {
		return Date{}, fmt.Errorf("%w: %s has no day %d", ErrDate, time.Date(y, m, 1, 0, 0, 0, 0, time.UTC).Format("January 2006"), d)
	}
	return Date{y, m, d}, nil
}

func dateOf(t time.Time) Date { return Date{t.Year(), t.Month(), t.Day()} }

// ParseClock reads a time of day: "3:30 pm", "3:30pm", "3:30 p.m.", "3pm",
// "15:30", "3:30:15 pm", "noon" or "midnight", optionally followed by a zone
// abbreviation such as "CST".
func ParseClock(s string) (Clock, error) {
	bad := fmt.Errorf("%w: %q", ErrClock, strings.TrimSpace(s))
	f := fields(s)
	if len(f) == 0 {
		return Clock{}, bad
	}
	var c Clock
	// A trailing zone abbreviation.
	if last := f[len(f)-1]; len(f) > 1 && isZone(last) {
		c.Zone = strings.ToUpper(last)
		f = f[:len(f)-1]
	}
	switch strings.Join(f, " ") {
	case "noon":
		c.Hour = 12
		return c, nil
	case "midnight":
		return c, nil
	}

	// Meridiem: "pm", "p m" (from "p.m."), or "p".
	meridiem := ""
	for len(f) > 0 {
		last := f[len(f)-1]
		if last != "am" && last != "pm" && last != "a" && last != "p" && last != "m" {
			break
		}
		meridiem = last + meridiem
		f = f[:len(f)-1]
	}
	switch meridiem {
	case "", "am", "pm":
	case "a", "p":
		meridiem += "m"
	default:
		return Clock{}, bad
	}

	if len(f) < 1 || len(f) > 3 {
		return Clock{}, bad
	}
	nums := make([]int, 3)
	for i, x := range f {
		n, err := strconv.Atoi(x)
		if err != nil || len(x) > 2 {
			return Clock{}, bad
		}
		nums[i] = n
	}
	c.Hour, c.Minute, c.Second = nums[0], nums[1], nums[2]
	if c.Minute > 59 || c.Second > 59 {
		return Clock{}, bad
	}
	switch meridiem {
	case "":
		if c.Hour > 23 || len(f) == 1 { // "3" alone is ambiguous
			return Clock{}, bad
		}
	default:
		if c.Hour < 1 || c.Hour > 12 {
			return Clock{}, bad
		}
		c.Hour %= 12
		if meridiem == "pm" {
			c.Hour += 12
		}
	}
	return c, nil
}

// zoneAbbreviations are the zones a time may name: the US zones, standard
// and daylight, and UTC.
var zoneAbbreviations = map[string]bool{
	"est": true, "edt": true, "cst": true, "cdt": true, "mst": true, "mdt": true, "pst": true, "pdt": true,
	"akst": true, "akdt": true, "hst": true, "hdt": true, "utc": true, "gmt": true,
}

func isZone(s string) bool { return zoneAbbreviations[s] }

// ErrNoSuchTime is returned for a local time that didn't occur, in the gap
// when clocks went forward.
var ErrNoSuchTime = errors.New("no such time")

// GapError reports a local time in the gap when clocks went forward.
type GapError struct {
	Local time.Time // the requested wall clock, in UTC for formatting
	// Later is the same moment by the clocks after the change: 2:30 am CST
	// is 3:30 am CDT.
	Later time.Time
}

func (e *GapError) Error() string {
	return fmt.Sprintf("%s didn't occur on %s: clocks went forward, so it would be %s",
		e.Local.Format("3:04 pm"), e.Local.Format("2 January 2006"), e.Later.Format("3:04 pm MST"))
}

func (e *GapError) Unwrap() error { return ErrNoSuchTime }

// ZoneError reports a zone abbreviation that wasn't in effect at a time.
type ZoneError struct {
	Zone string   // as given
	Want []string // the zones that were in effect
}

func (e *ZoneError) Error() string {
	return fmt.Sprintf("%s wasn't in effect then; it was %s", e.Zone, strings.Join(e.Want, " or "))
}

// Resolve returns the instants a local date and time can mean in loc,
// earliest first: one normally; two when clocks went back and the time
// happened twice, once in each zone (CDT, then CST); and none, with a
// *GapError, when clocks went forward past it. A zone abbreviation on the
// clock picks between two.
func Resolve(d Date, c Clock, loc *time.Location) ([]time.Time, error) {
	wall := time.Date(d.Year, d.Month, d.Day, c.Hour, c.Minute, c.Second, 0, time.UTC)

	// The offsets in effect around the date, and the instants each would
	// give; keep those that really show this wall clock.
	var offsets []int
	for _, h := range []int{-36, -12, 0, 12, 36} {
		_, off := wall.Add(time.Duration(h) * time.Hour).In(loc).Zone()
		if !slices.Contains(offsets, off) {
			offsets = append(offsets, off)
		}
	}
	var out []time.Time
	for _, off := range offsets {
		t := wall.Add(-time.Duration(off) * time.Second).In(loc)
		if y, m, dd := t.Date(); y == d.Year && m == d.Month && dd == d.Day &&
			t.Hour() == c.Hour && t.Minute() == c.Minute && t.Second() == c.Second && !slices.ContainsFunc(out, t.Equal) {
			out = append(out, t)
		}
	}
	slices.SortFunc(out, func(a, b time.Time) int { return a.Compare(b) })

	if len(out) == 0 {
		// Read the wall clock by the offset before the change.
		_, before := wall.Add(-36 * time.Hour).In(loc).Zone()
		return nil, &GapError{Local: wall, Later: wall.Add(-time.Duration(before) * time.Second).In(loc)}
	}
	if c.Zone != "" {
		var zones []string
		var kept []time.Time
		for _, t := range out {
			z, _ := t.Zone()
			zones = append(zones, z)
			if z == c.Zone {
				kept = append(kept, t)
			}
		}
		if len(kept) == 0 {
			return nil, &ZoneError{Zone: c.Zone, Want: zones}
		}
		out = kept
	}
	return out, nil
}
