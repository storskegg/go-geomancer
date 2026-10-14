package input

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/storskegg/go-geomancer/figure"
	"github.com/storskegg/go-geomancer/system"
	"github.com/storskegg/go-geomancer/zodiac"
)

func TestPromptAccessible(t *testing.T) {
	var out bytes.Buffer
	p := Prompt{In: strings.NewReader("6\n8\n1\n2\n"), Out: &out, Accessible: true}
	got, err := p.Mothers()
	if err != nil {
		t.Fatal(err)
	}
	want := [4]figure.Figure{figure.Amissio, figure.FortunaMinor, figure.Via, figure.Populus}
	if got != want {
		t.Errorf("Mothers() = %v, want %v", got, want)
	}
	if !strings.Contains(out.String(), "Mother 4 (M4)") {
		t.Errorf("output lacks the fourth prompt:\n%s", out.String())
	}
}

func TestPromptAccessibleSystem(t *testing.T) {
	a := testSystem("a", "Alpha")
	b := testSystem("b", "Beta")
	chosen := system.System(b)
	var out bytes.Buffer
	p := Prompt{
		In: strings.NewReader("1\n6\n8\n1\n2\n"), Out: &out, Accessible: true,
		Systems: []system.System{a, b}, System: &chosen,
	}
	got, err := p.Mothers()
	if err != nil {
		t.Fatal(err)
	}
	if chosen != a {
		t.Errorf("system = %v, want Alpha", chosen.Name())
	}
	if got[3] != figure.Populus {
		t.Errorf("Mothers() = %v", got)
	}
	if !strings.Contains(out.String(), "1. Alpha") {
		t.Errorf("output lacks the system prompt:\n%s", out.String())
	}
}

func TestDescribeSystem(t *testing.T) {
	got := describeSystem(testSystem("a", "Alpha"))
	want := "Head Fire · Neck Air · Body Water · Feet Earth\nMothers in houses 1, 2, 3, 4"
	if got != want {
		t.Errorf("describeSystem =\n%s\nwant\n%s", got, want)
	}
}

func testSystem(id, name string) *system.Mapping {
	return &system.Mapping{
		Slug:            id,
		Title:           name,
		SegmentElements: [4]zodiac.Element{zodiac.Fire, zodiac.Air, zodiac.Water, zodiac.Earth},
		Houses:          [12]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
	}
}

// driver feeds key presses to a form and runs the commands it returns, so
// the interactive select can be tested without a terminal.
type driver struct {
	t    *testing.T
	form *huh.Form
}

func newDriver(t *testing.T, m *[4]figure.Figure) *driver {
	d := &driver{t: t, form: newForm(motherGroups(m, false)...)}
	d.run(d.form.Init())
	return d
}

func (d *driver) send(msg tea.Msg) {
	_, cmd := d.form.Update(msg)
	d.run(cmd)
}

// run executes cmd and feeds back its messages, skipping commands that
// don't finish promptly (such as cursor blink timers).
func (d *driver) run(cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	var msg tea.Msg
	select {
	case msg = <-done:
	case <-time.After(50 * time.Millisecond):
		return
	}
	switch msg := msg.(type) {
	case nil:
	case tea.BatchMsg:
		for _, c := range msg {
			d.run(c)
		}
	case tea.QuitMsg:
	default:
		d.send(msg)
	}
}

func (d *driver) typeText(s string) {
	for _, r := range s {
		d.send(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

func (d *driver) key(code rune) { d.send(tea.KeyPressMsg{Code: code}) }

func (d *driver) view() string { return ansi.Strip(d.form.View()) }

func TestPromptInteractive(t *testing.T) {
	var m [4]figure.Figure
	d := newDriver(t, &m)

	if v := d.view(); !strings.Contains(v, "Mother 1 (M1)") || !strings.Contains(v, "Via") {
		t.Fatalf("initial view:\n%s", v)
	}

	// Typing filters by name.
	d.typeText("fortuna")
	v := d.view()
	if !strings.Contains(v, "Fortuna Major") || !strings.Contains(v, "Fortuna Minor") || strings.Contains(v, "Populus") {
		t.Errorf("view after typing \"fortuna\":\n%s", v)
	}
	d.key(tea.KeyDown)
	if m[0] != figure.FortunaMinor {
		t.Errorf("after ↓, M1 = %v, want Fortuna Minor", m[0])
	}
	if v := d.view(); !strings.Contains(v, "Sun") {
		t.Errorf("preview doesn't follow the cursor:\n%s", v)
	}
	d.key(tea.KeyEnter)

	// Typing filters by pattern too, and the preview follows the filter.
	if v := d.view(); !strings.Contains(v, "Mother 2 (M2)") {
		t.Fatalf("second prompt not shown:\n%s", v)
	}
	d.typeText("1212")
	if v := d.view(); !strings.Contains(v, "/1212") || !strings.Contains(v, "Venus") {
		t.Errorf("view after typing \"1212\" lacks the filter or Amissio's preview:\n%s", v)
	}
	d.key(tea.KeyEnter)

	d.key(tea.KeyEnter) // M3: the first figure, Via
	d.typeText("pop")
	d.key(tea.KeyEnter)

	want := [4]figure.Figure{figure.FortunaMinor, figure.Amissio, figure.Via, figure.Populus}
	if m != want {
		t.Errorf("Mothers = %v, want %v", m, want)
	}
	if d.form.State != huh.StateCompleted {
		t.Errorf("form state = %v, want completed", d.form.State)
	}
}
