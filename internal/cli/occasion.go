package cli

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/storskegg/go-geomancer/internal/cli/input"
	"github.com/storskegg/go-geomancer/place"
	"github.com/storskegg/go-geomancer/when"
)

// occasionFlags are --date, --time and --place, as given.
type occasionFlags struct{ date, time, place string }

func (f occasionFlags) any() bool { return f.date != "" || f.time != "" || f.place != "" }

// asks says which parts of the occasion still need asking for.
type asks struct{ date, time, place bool }

// readOccasion reads --date, --time and --place. With prompts (interactive),
// anything missing is asked for, and anything that can't be read is
// reported on warn and asked for too. Without prompts, the three go
// together: none gives a nil occasion, and all three must be readable.
func readOccasion(f occasionFlags, interactive bool, now time.Time, warn io.Writer) (*input.Occasion, asks, error) {
	if !interactive {
		if !f.any() {
			return nil, asks{}, nil
		}
		if f.date == "" || f.time == "" || f.place == "" {
			return nil, asks{}, errors.New("--date, --time and --place go together: give all three, or none")
		}
	}
	o := &input.Occasion{}
	need := asks{date: f.date == "", time: f.time == "", place: f.place == ""}
	problem := func(flag, value string, err error) error {
		if !interactive {
			return fmt.Errorf("--%s %q: %w", flag, value, err)
		}
		_, _ = fmt.Fprintf(warn, "Couldn't use --%s %q (%v); asking instead.\n", flag, value, err)
		return nil
	}

	if f.date != "" {
		d, err := when.ParseDate(f.date, now)
		if err == nil {
			o.Date, o.DateText = d, f.date
		} else if err := problem("date", f.date, err); err != nil {
			return nil, asks{}, err
		} else {
			need.date = true
		}
	}
	if f.time != "" {
		c, err := when.ParseClock(f.time)
		if err == nil {
			o.Clock, o.TimeText = c, f.time
		} else if err := problem("time", f.time, err); err != nil {
			return nil, asks{}, err
		} else {
			need.time = true
		}
	}
	if f.place != "" {
		m, err := findPlace(f.place)
		if err == nil {
			o.Place, o.PlaceText = m, m.Label
		} else if err := problem("place", f.place, err); err != nil {
			return nil, asks{}, err
		} else {
			need.place = true
		}
	}

	// With everything given, resolve the time now. A time that didn't occur
	// is asked for again; one that happened twice is asked about, or without
	// prompts must say which zone ("1:30 am CST").
	if !need.date && !need.time && !need.place {
		if err := o.Resolve(); err != nil {
			if err := problem("time", f.time, err); err != nil {
				return nil, asks{}, err
			}
			need.time = true
		} else if o.Ambiguous() && !interactive {
			z0, _ := o.Candidates[0].Zone()
			z1, _ := o.Candidates[1].Zone()
			return nil, asks{}, fmt.Errorf("--time %q happened twice at %s on %s, when the clocks went back: add %s or %s",
				f.time, o.Place.Label, o.Date, z0, z1)
		}
	}
	return o, need, nil
}

// findPlace finds a place as written: its label, or failing that the best
// match.
func findPlace(s string) (place.Match, error) {
	ix, err := place.Default()
	if err != nil {
		return place.Match{}, err
	}
	if m, ok := ix.Find(s); ok {
		return m, nil
	}
	if ms := ix.Search(s, 1); len(ms) > 0 {
		return ms[0], nil
	}
	return place.Match{}, errors.New("no place by that name")
}

// places returns the place index for prompts, which need it only once the
// querent starts typing a place.
func places() *place.Index {
	ix, err := place.Default()
	if err != nil {
		panic(err) // the embedded data is checked by the package tests
	}
	return ix
}
