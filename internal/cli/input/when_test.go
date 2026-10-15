package input

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/place"
)

var testNow = func() time.Time { return time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC) }

func testPlaces() *place.Index {
	ix, err := place.Default()
	if err != nil {
		panic(err)
	}
	return ix
}

func occasionDriver(t *testing.T, o *Occasion) *driver {
	return newDriver(t, dateGroup(o, testNow, false), timeGroup(o, false), placeGroup(o, testPlaces, false), zoneGroup(o))
}

func TestOccasionPrompts(t *testing.T) {
	o := &Occasion{}
	d := occasionDriver(t, o)
	if v := d.view(); !strings.Contains(v, "Date") || !strings.Contains(v, "Such as 15 October 2026") {
		t.Fatalf("date prompt:\n%s", v)
	}

	// The date is described as it's typed, and must be readable.
	d.typeText("10/15/20")
	if v := d.view(); !strings.Contains(v, "→ Thursday, 15 October 2020") {
		t.Errorf("partial date:\n%s", v)
	}
	d.typeText("26")
	if v := d.view(); !strings.Contains(v, "→ Thursday, 15 October 2026") {
		t.Errorf("full date:\n%s", v)
	}
	d.key(tea.KeyEnter)

	d.typeText("half past")
	d.key(tea.KeyEnter)
	if v := d.view(); !strings.Contains(v, "Not a time I can read") {
		t.Errorf("bad time accepted:\n%s", v)
	}
	for range "half past" {
		d.key(tea.KeyBackspace)
	}
	d.typeText("3:30 pm")
	if v := d.view(); !strings.Contains(v, "→ 3:30 pm") {
		t.Errorf("time preview:\n%s", v)
	}
	d.key(tea.KeyEnter)

	// The place lists matches as it's typed; ↓ and Enter take one.
	d.typeText("springf")
	v := d.view()
	if !strings.Contains(v, "> Springfield, MO") || !strings.Contains(v, "  Springfield, MA") {
		t.Fatalf("place matches:\n%s", v)
	}
	d.key(tea.KeyDown)
	d.key(tea.KeyDown)
	if v := d.view(); !strings.Contains(v, "> Springfield, IL") {
		t.Errorf("after ↓↓:\n%s", v)
	}
	d.key(tea.KeyEnter)

	if d.form.State != huh.StateCompleted {
		t.Fatalf("form not completed (the zone question should be hidden):\n%s", d.view())
	}
	if got := o.Time().Format("2006-01-02 15:04 MST"); got != "2026-10-15 15:30 CDT" || o.Place.Label != "Springfield, IL" {
		t.Errorf("occasion = %s at %s", got, o.Place.Label)
	}
}

// A time that didn't occur at the place stops at the place, with a reason.
func TestOccasionGap(t *testing.T) {
	o := &Occasion{}
	d := occasionDriver(t, o)
	d.typeText("March 8, 2026")
	d.key(tea.KeyEnter)
	d.typeText("2:30 am")
	d.key(tea.KeyEnter)
	d.typeText("chicago il")
	d.key(tea.KeyEnter)
	v := d.view()
	if !strings.Contains(v, "didn't occur on 8 March 2026") || !strings.Contains(v, "3:30 am CDT") || d.form.State == huh.StateCompleted {
		t.Fatalf("gap not reported:\n%s", v)
	}

	// Going back and changing the time checks it against the place.
	d.send(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	for range "2:30 am" {
		d.key(tea.KeyBackspace)
	}
	d.typeText("3:30 am")
	d.key(tea.KeyEnter)
	d.key(tea.KeyEnter) // the place, as before
	if d.form.State != huh.StateCompleted || o.Time().Format("15:04 MST") != "03:30 CDT" {
		t.Errorf("state %v, time %v:\n%s", d.form.State, o.Time(), d.view())
	}
}

// A time that happened twice asks which.
func TestOccasionOverlap(t *testing.T) {
	o := &Occasion{}
	d := occasionDriver(t, o)
	d.typeText("2026-11-01")
	d.key(tea.KeyEnter)
	d.typeText("1:30 am")
	d.key(tea.KeyEnter)
	d.typeText("chicago il")
	d.key(tea.KeyEnter)
	v := d.view()
	if !strings.Contains(v, "Which time?") || !strings.Contains(v, "1:30 am CDT (the first time") || !strings.Contains(v, "1:30 am CST (the second time") {
		t.Fatalf("zone question:\n%s", v)
	}
	d.key(tea.KeyDown)
	d.key(tea.KeyEnter)
	if d.form.State != huh.StateCompleted || o.Time().Format("15:04 MST") != "01:30 CST" {
		t.Errorf("state %v, time %v", d.form.State, o.Time())
	}
}

func TestOccasionNoPlace(t *testing.T) {
	o := &Occasion{}
	d := newDriver(t, placeGroup(o, testPlaces, false))
	d.typeText("zzzzqq")
	if v := d.view(); !strings.Contains(v, "No matches") {
		t.Errorf("view:\n%s", v)
	}
	d.key(tea.KeyEnter)
	if v := d.view(); !strings.Contains(v, "No place by that name") {
		t.Errorf("unknown place accepted:\n%s", v)
	}
}

func TestPromptAccessibleOccasion(t *testing.T) {
	o := &Occasion{}
	var out bytes.Buffer
	p := Prompt{
		// Date, time, place, which time (2: CST), then the Mothers.
		In: strings.NewReader("11/1/2026\n1:30 am\nChicago, IL\n2\n6\n8\n1\n2\n"), Out: &out, Accessible: true,
		Occasion: o, AskDate: true, AskTime: true, AskPlace: true, Places: testPlaces, Now: testNow,
	}
	got, err := p.Mothers()
	if err != nil {
		t.Fatal(err)
	}
	if o.Time().Format("2006-01-02 15:04 MST") != "2026-11-01 01:30 CST" || got[0] != figure.Amissio || got[3] != figure.Populus {
		t.Errorf("time %v, mothers %v", o.Time(), got)
	}
	if !strings.Contains(out.String(), "Which time?") {
		t.Errorf("no zone question:\n%s", out.String())
	}

	// Without an ambiguous time, there's no zone question.
	o = &Occasion{}
	out.Reset()
	p.Occasion, p.In = o, strings.NewReader("today\nnoon\nspringfield il\n6\n8\n1\n2\n")
	if _, err := p.Mothers(); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "Which time?") || o.Time().Format("2006-01-02 15:04 MST") != "2026-10-15 12:00 CDT" {
		t.Errorf("time %v:\n%s", o.Time(), out.String())
	}
}

func TestOccasionTime(t *testing.T) {
	o := &Occasion{}
	if !o.Time().IsZero() || o.Ambiguous() {
		t.Error("an unresolved occasion has a time")
	}
	if err := o.recheck(); err != nil {
		t.Errorf("recheck without a place = %v", err)
	}
}
