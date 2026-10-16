package render

import (
	"strings"
	"testing"
	"time"

	"github.com/storskegg/go-geomancer/sky"
)

func skyAt(t *testing.T, tm time.Time, lat, lon float64) *sky.Sky {
	t.Helper()
	s, err := sky.At(tm, lat, lon)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestHeavens(t *testing.T) {
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		t.Fatal(err)
	}
	s := skyAt(t, time.Date(2026, 10, 15, 13, 14, 0, 0, loc), 43.0731, -89.4012)

	out := plain(Heavens(s, Options{}))
	for _, want := range []string{
		"Day       ♃ Jupiter · Thursday, from sunrise at 7:11 am",
		"Hour      ♄ Saturn · 7th hour of the day, 12:43 pm–1:38 pm",
		"Moon      Crescent · waxing · 24% lit",
		"│ ♀ Venus      │ 5°28′ Scorpio      │ Retrograde │ Detriment · Under the beams │",
		"│ ☊ North Node │ 26°55′ Aquarius    │ —          │ —                           │",
		"│ Mars △ Saturn     │ 0°22′  │ applying   │",
		"│ Previous │ Sat 10 Oct 2026, 10:49 am CDT │ Sat 26 Sep 2026, 11:48 am CDT │",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}

	ascii := plain(Heavens(s, Options{Symbols: ASCII}))
	for _, want := range []string{"Day       Jupiter ·", "│ Mars trine Saturn", "│ Sun        │"} {
		if !strings.Contains(ascii, want) {
			t.Errorf("ASCII output lacks %q:\n%s", want, ascii)
		}
	}
	if strings.ContainsAny(ascii, "☉☽☿♀♂♃♄☊☋☌⚹□△☍") {
		t.Errorf("ASCII output has symbols:\n%s", ascii)
	}
}

func TestHeavensNoHour(t *testing.T) {
	loc, err := time.LoadLocation("America/Anchorage")
	if err != nil {
		t.Fatal(err)
	}
	out := plain(Heavens(skyAt(t, time.Date(2026, 6, 21, 12, 0, 0, 0, loc), 71.2906, -156.7886), Options{}))
	for _, want := range []string{"Day       ☉ Sun · Sunday", "Hour      not available: the Sun doesn't set here on this date"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}
