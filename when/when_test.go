package when

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var now = time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)

func TestParseDate(t *testing.T) {
	oct15 := Date{2026, time.October, 15}
	tests := map[string]Date{
		"2026-10-15":                 oct15,
		"2026/10/15":                 oct15,
		"10/15/2026":                 oct15,
		"10/15/26":                   oct15,
		"10-15-2026":                 oct15,
		"Oct 15 2026":                oct15,
		"oct. 15, 2026":              oct15,
		"October 15, 2026":           oct15,
		"15 October 2026":            oct15,
		"15 Oct 2026":                oct15,
		"Thursday, October 15, 2026": oct15,
		"Thu 15 Oct 2026":            oct15,
		"Sept 1 2026":                {2026, time.September, 1},
		"today":                      oct15,
		" Yesterday ":                {2026, time.October, 14},
		"tomorrow":                   {2026, time.October, 16},
		"2/29/2028":                  {2028, time.February, 29}, // a leap year
		"1/1/1900":                   {1900, time.January, 1},
	}
	for in, want := range tests {
		got, err := ParseDate(in, now)
		if err != nil || got != want {
			t.Errorf("ParseDate(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
}

func TestParseDateInvalid(t *testing.T) {
	for _, in := range []string{"", "soon", "2026-13-01", "2/29/2026", "4/31/2026", "Octember 5 2026", "10/15", "15 2026", "2026-10-15-01", "0/10/2026"} {
		if _, err := ParseDate(in, now); !errors.Is(err, ErrDate) {
			t.Errorf("ParseDate(%q) error = %v, want ErrDate", in, err)
		}
	}
	_, err := ParseDate("4/31/2026", now)
	if err == nil || !strings.Contains(err.Error(), "April 2026 has no day 31") {
		t.Errorf("31 April: %v", err)
	}
}

func TestParseClock(t *testing.T) {
	tests := map[string]Clock{
		"3:30 pm":       {Hour: 15, Minute: 30},
		"3:30pm":        {Hour: 15, Minute: 30},
		"3:30 p.m.":     {Hour: 15, Minute: 30},
		"3:30 P.M.":     {Hour: 15, Minute: 30},
		"3pm":           {Hour: 15},
		"3 p":           {Hour: 15},
		"12:15 am":      {Minute: 15},
		"12 pm":         {Hour: 12},
		"15:30":         {Hour: 15, Minute: 30},
		"00:05":         {Minute: 5},
		"3:30:15 pm":    {Hour: 15, Minute: 30, Second: 15},
		"noon":          {Hour: 12},
		"Midnight":      {},
		"1:30 am CST":   {Hour: 1, Minute: 30, Zone: "CST"},
		"1:30 a.m. cdt": {Hour: 1, Minute: 30, Zone: "CDT"},
		"13:00 EST":     {Hour: 13, Zone: "EST"},
	}
	for in, want := range tests {
		got, err := ParseClock(in)
		if err != nil || got != want {
			t.Errorf("ParseClock(%q) = %+v, %v; want %+v", in, got, err, want)
		}
	}
}

func TestParseClockInvalid(t *testing.T) {
	for _, in := range []string{"", "3", "13 pm", "0 am", "24:00", "3:60 pm", "3:30:60", "half past three", "3:30 xm", "123:00"} {
		if _, err := ParseClock(in); !errors.Is(err, ErrClock) {
			t.Errorf("ParseClock(%q) error = %v, want ErrClock", in, err)
		}
	}
}

func TestStrings(t *testing.T) {
	if got := (Date{2026, time.October, 15}).String(); got != "Thursday, 15 October 2026" {
		t.Errorf("Date.String() = %q", got)
	}
	for c, want := range map[Clock]string{
		{Hour: 15, Minute: 30}:             "3:30 pm",
		{Hour: 0, Minute: 5, Second: 9}:    "12:05:09 am",
		{Hour: 1, Minute: 30, Zone: "CST"}: "1:30 am CST",
	} {
		if got := c.String(); got != want {
			t.Errorf("%+v.String() = %q, want %q", c, got, want)
		}
	}
}

func load(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func zones(ts []time.Time) string {
	var z []string
	for _, t := range ts {
		s, _ := t.Zone()
		z = append(z, t.Format("15:04 ")+s)
	}
	return strings.Join(z, ", ")
}

func TestResolve(t *testing.T) {
	chicago := load(t, "America/Chicago")
	tests := []struct {
		d    Date
		c    Clock
		loc  *time.Location
		want string // instants, as "15:04 MST", comma-separated
	}{
		{Date{2026, time.October, 15}, Clock{Hour: 15, Minute: 30}, chicago, "15:30 CDT"},
		{Date{2026, time.January, 15}, Clock{Hour: 15, Minute: 30}, chicago, "15:30 CST"},
		// Clocks go back at 2:00 CDT on 1 November 2026: 1:30 happens twice.
		{Date{2026, time.November, 1}, Clock{Hour: 1, Minute: 30}, chicago, "01:30 CDT, 01:30 CST"},
		{Date{2026, time.November, 1}, Clock{Hour: 1, Minute: 30, Zone: "CST"}, chicago, "01:30 CST"},
		{Date{2026, time.November, 1}, Clock{Hour: 1, Minute: 30, Zone: "CDT"}, chicago, "01:30 CDT"},
		{Date{2026, time.November, 1}, Clock{Hour: 0, Minute: 59}, chicago, "00:59 CDT"},
		{Date{2026, time.November, 1}, Clock{Hour: 2}, chicago, "02:00 CST"},
		// Clocks go forward at 2:00 CST on 8 March 2026.
		{Date{2026, time.March, 8}, Clock{Hour: 3}, chicago, "03:00 CDT"},
		{Date{2026, time.March, 8}, Clock{Hour: 1, Minute: 59}, chicago, "01:59 CST"},
		// Arizona keeps standard time all year.
		{Date{2026, time.July, 4}, Clock{Hour: 12}, load(t, "America/Phoenix"), "12:00 MST"},
		{Date{2026, time.July, 4}, Clock{Hour: 12}, load(t, "Pacific/Honolulu"), "12:00 HST"},
	}
	for _, tt := range tests {
		got, err := Resolve(tt.d, tt.c, tt.loc)
		if err != nil || zones(got) != tt.want {
			t.Errorf("Resolve(%v, %v, %v) = %s, %v; want %s", tt.d, tt.c, tt.loc, zones(got), err, tt.want)
		}
	}
}

func TestResolveGap(t *testing.T) {
	chicago := load(t, "America/Chicago")
	_, err := Resolve(Date{2026, time.March, 8}, Clock{Hour: 2, Minute: 30}, chicago)
	var gap *GapError
	if !errors.As(err, &gap) || !errors.Is(err, ErrNoSuchTime) {
		t.Fatalf("error = %v, want a GapError", err)
	}
	if got := gap.Later.Format("15:04 MST"); got != "03:30 CDT" {
		t.Errorf("Later = %s, want 03:30 CDT", got)
	}
	if want := "2:30 am didn't occur on 8 March 2026: clocks went forward, so it would be 3:30 am CDT"; err.Error() != want {
		t.Errorf("error = %q, want %q", err, want)
	}
}

func TestResolveWrongZone(t *testing.T) {
	_, err := Resolve(Date{2026, time.October, 15}, Clock{Hour: 15, Zone: "CST"}, load(t, "America/Chicago"))
	var ze *ZoneError
	if !errors.As(err, &ze) || err.Error() != "CST wasn't in effect then; it was CDT" {
		t.Errorf("error = %v", err)
	}
}
