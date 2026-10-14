package input

import (
	"bytes"
	"errors"
	"fmt"
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
		Planets: map[figure.Figure]zodiac.Planet{
			figure.FortunaMinor: zodiac.Sun, figure.Amissio: zodiac.Venus, figure.Puer: zodiac.Mars,
		},
	}
}

// driver feeds key presses to a form and runs the commands it returns, so
// the interactive select can be tested without a terminal.
type driver struct {
	t    *testing.T
	form *huh.Form
}

func newDriver(t *testing.T, groups ...*huh.Group) *driver {
	d := &driver{t: t, form: newForm(groups...)}
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
	sys := system.System(testSystem("a", "Alpha"))
	d := newDriver(t, motherGroups(&m, &sys, false)...)

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

func TestSystemSelect(t *testing.T) {
	a, b := testSystem("a", "Alpha"), testSystem("b", "Beta")
	b.SegmentElements = [4]zodiac.Element{zodiac.Fire, zodiac.Water, zodiac.Air, zodiac.Earth}
	var chosen system.System // nil: defaults to the first system
	d := newDriver(t, systemGroup(&chosen, []system.System{a, b}, false))

	if chosen != system.System(a) {
		t.Fatalf("initial choice = %v, want Alpha", chosen)
	}
	v := d.view()
	if !strings.Contains(v, "System") || !strings.Contains(v, "Beta") || !strings.Contains(v, "Neck Air") {
		t.Errorf("initial view:\n%s", v)
	}

	d.key(tea.KeyDown)
	if chosen != system.System(b) {
		t.Errorf("after ↓, choice = %v, want Beta", chosen)
	}
	if v := d.view(); !strings.Contains(v, "Neck Water") {
		t.Errorf("description doesn't follow the cursor:\n%s", v)
	}
	d.key(tea.KeyEnter)
	if d.form.State != huh.StateCompleted || chosen != system.System(b) {
		t.Errorf("state = %v, choice = %v", d.form.State, chosen)
	}
}

func TestSystemSelectKeepsInitialChoice(t *testing.T) {
	a, b := testSystem("a", "Alpha"), testSystem("b", "Beta")
	chosen := system.System(b)
	d := newDriver(t, systemGroup(&chosen, []system.System{a, b}, false))
	d.key(tea.KeyEnter)
	if chosen != system.System(b) {
		t.Errorf("choice = %v, want Beta", chosen)
	}
}

// Shift+Tab from the first Mother returns to the system.
func TestPromptGoBack(t *testing.T) {
	a, b := testSystem("a", "Alpha"), testSystem("b", "Beta")
	chosen := system.System(a)
	var m [4]figure.Figure
	d := newDriver(t, append([]*huh.Group{systemGroup(&chosen, []system.System{a, b}, false)}, motherGroups(&m, &chosen, false)...)...)

	d.key(tea.KeyEnter)
	if v := d.view(); !strings.Contains(v, "Mother 1 (M1)") {
		t.Fatalf("not on Mother 1:\n%s", v)
	}
	d.send(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	if v := d.view(); !strings.Contains(v, "Mothers in houses") {
		t.Fatalf("shift+tab didn't return to the system:\n%s", v)
	}
	d.key(tea.KeyDown)
	d.key(tea.KeyEnter)
	for range 4 {
		d.key(tea.KeyEnter)
	}
	if chosen != system.System(b) || d.form.State != huh.StateCompleted {
		t.Errorf("choice = %v, state = %v", chosen.Name(), d.form.State)
	}
}

func TestPromptAbort(t *testing.T) {
	var m [4]figure.Figure
	d := newDriver(t, motherGroups(&m, nil, false)...)
	d.send(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if d.form.State != huh.StateAborted {
		t.Errorf("state after ctrl+c = %v, want aborted", d.form.State)
	}
}

func TestFormError(t *testing.T) {
	other := errors.New("boom")
	tests := []struct{ in, want error }{
		{nil, nil},
		{huh.ErrUserAborted, ErrAborted},
		{fmt.Errorf("wrapped: %w", huh.ErrUserAborted), ErrAborted},
		{other, other},
	}
	for _, tt := range tests {
		if got := formError(tt.in); !errors.Is(got, tt.want) || (tt.want == nil && got != nil) {
			t.Errorf("formError(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestByteReader(t *testing.T) {
	r := byteReader{strings.NewReader("abc")}
	buf := make([]byte, 8)
	var got []byte
	for {
		n, err := r.Read(buf)
		if n > 1 {
			t.Fatalf("Read returned %d bytes, want at most 1", n)
		}
		got = append(got, buf[:n]...)
		if err != nil {
			break
		}
	}
	if string(got) != "abc" {
		t.Errorf("read %q, want abc", got)
	}
}

func TestPreview(t *testing.T) {
	got := ansi.Strip(preview(figure.Puer, testSystem("a", "Alpha"))) // 1121
	want := strings.Join([]string{
		"  ●     Puer",
		"  ●     1121",
		"●   ●   Mars",
		"  ●     ",
	}, "\n")
	if got != want {
		t.Errorf("preview =\n%q\nwant\n%q", got, want)
	}

	// Without a system, or with one that gives no planet, the line is blank.
	for _, sys := range []system.System{nil, testSystem("a", "Alpha")} {
		if got := ansi.Strip(preview(figure.Via, sys)); strings.Split(got, "\n")[2] != "  ●     " {
			t.Errorf("preview(Via) planet line = %q, want blank", strings.Split(got, "\n")[2])
		}
	}
}

// Going back and choosing another system updates the preview's planet.
func TestPreviewFollowsSystem(t *testing.T) {
	a, b := testSystem("a", "Alpha"), testSystem("b", "Beta")
	b.Planets = map[figure.Figure]zodiac.Planet{figure.Via: zodiac.Saturn}
	chosen := system.System(a)
	var m [4]figure.Figure
	d := newDriver(t, append([]*huh.Group{systemGroup(&chosen, []system.System{a, b}, false)}, motherGroups(&m, &chosen, false)...)...)

	d.key(tea.KeyEnter) // Alpha; M1 starts on Via, which Alpha gives no planet
	if v := d.view(); strings.Contains(v, "Saturn") {
		t.Fatalf("Alpha preview shows Beta's planet:\n%s", v)
	}
	d.send(tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	d.key(tea.KeyDown) // Beta
	d.key(tea.KeyEnter)
	if v := d.view(); !strings.Contains(v, "Mother 1 (M1)") || !strings.Contains(v, "Saturn") {
		t.Errorf("preview after switching to Beta:\n%s", v)
	}
}
