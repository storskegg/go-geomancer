package chart

import (
	"strings"
	"testing"
	"time"

	"github.com/storskegg/go-geomancer/sky"
	"github.com/storskegg/go-geomancer/zodiac"
)

// testSky is a sky with the hour and day of the given rulers, a waxing
// Moon in Sagittarius, peregrine and far from the Sun, and the Moon's next
// aspect as given (nil: void of course).
func testSky(hour, day zodiac.Planet, next *sky.MoonAspect) *sky.Sky {
	at := time.Date(2026, 10, 15, 15, 57, 0, 0, time.UTC)
	s := &sky.Sky{
		Day:  sky.Day{Ruler: day, Hour: &sky.Hour{Ruler: hour, Number: 7}},
		Moon: sky.Moon{Phase: sky.Crescent, Illuminated: 0.25, LeavesSign: at.Add(10 * time.Hour), NextAspect: next, VoidOfCourse: next == nil},
	}
	for _, p := range zodiac.Planets() {
		s.Bodies = append(s.Bodies, sky.Body{Planet: p, Sign: zodiac.Sagittarius, Dignities: []sky.Dignity{sky.Peregrine}})
	}
	return s
}

// moonJudgement finds the judgement of the Moon containing text.
func moonJudgement(js []Judgement, contains string) (Judgement, bool) {
	for _, j := range js {
		if j.About == "Moon" && strings.Contains(j.Text, contains) {
			return j, true
		}
	}
	return Judgement{}, false
}

func TestTimingHourAndDay(t *testing.T) {
	venus := []zodiac.Planet{zodiac.Venus}
	tests := []struct {
		hour   zodiac.Planet
		effect Effect
		text   string
	}{
		{zodiac.Venus, Helps, "The hour of Venus, natural significator of marriage, suits the question."},
		{zodiac.Jupiter, Helps, "The hour of Jupiter, the greater benefic, favours the question, though Venus signifies marriage."},
		{zodiac.Saturn, Hinders, "The hour of Saturn, the greater malefic, is unfortunate for marriage, which Venus signifies."},
		{zodiac.Mars, Hinders, "The hour of Mars, the lesser malefic, is unfortunate for marriage, which Venus signifies."},
		{zodiac.Sun, Neutral, "The hour of the Sun neither suits nor hinders marriage, which Venus signifies."},
	}
	for _, tt := range tests {
		js := Timing(testSky(tt.hour, zodiac.Mercury, nil), "marriage", venus)
		if js[0].About != "Hour" || js[0].Effect != tt.effect || js[0].Text != tt.text {
			t.Errorf("hour of %v: %+v, want %v %q", tt.hour, js[0], tt.effect, tt.text)
		}
	}

	// A matter of two planets, and the day.
	js := Timing(testSky(zodiac.Moon, zodiac.Mars, nil), "sickness", []zodiac.Planet{zodiac.Mars, zodiac.Saturn})
	if j := js[1]; j.About != "Day" || j.Effect != Helps || j.Text != "The day of Mars, natural significator of sickness, suits the question." {
		t.Errorf("day: %+v", j)
	}
	if j := js[0]; j.Text != "The hour of the Moon neither suits nor hinders sickness, which Mars and Saturn signify." {
		t.Errorf("hour: %+v", j)
	}

	// No planetary hour.
	s := testSky(zodiac.Sun, zodiac.Sun, nil)
	s.Day.Hour, s.Day.Unavailable = nil, "the Sun doesn't set here on this date"
	if j := Timing(s, "marriage", venus)[0]; j.Effect != Neutral || j.Text != "No planetary hour: the Sun doesn't set here on this date." {
		t.Errorf("no hour: %+v", j)
	}

	// No matter: only the Moon.
	for _, j := range Timing(testSky(zodiac.Venus, zodiac.Venus, nil), "", nil) {
		if j.About != "Moon" {
			t.Errorf("without a matter: %+v", j)
		}
	}
}

func TestTimingMoon(t *testing.T) {
	at := time.Date(2026, 10, 15, 20, 57, 0, 0, time.UTC)
	aspect := func(p zodiac.Planet, k sky.Kind) *sky.MoonAspect {
		return &sky.MoonAspect{Planet: p, Kind: k, Exact: at}
	}
	tests := []struct {
		next   *sky.MoonAspect
		effect Effect
		text   string
	}{
		{nil, Hinders, "The Moon is void of course: it makes no aspect before leaving Sagittarius at 1:57 am Friday."},
		{aspect(zodiac.Jupiter, sky.Trine), Helps, "The Moon's next aspect is a trine to Jupiter, exact at 8:57 pm Thursday. A benefic, well met: a good sign for the matter."},
		{aspect(zodiac.Venus, sky.Square), Mixed, "a square to Venus"},
		{aspect(zodiac.Saturn, sky.Sextile), Mixed, "A malefic, though by an easy aspect"},
		{aspect(zodiac.Mars, sky.Opposition), Hinders, "an opposition to Mars"},
		{aspect(zodiac.Saturn, sky.Conjunction), Hinders, "a conjunction with Saturn"},
		{aspect(zodiac.Mercury, sky.Trine), Helps, "An easy aspect"},
		{aspect(zodiac.Sun, sky.Square), Hinders, "a square to the Sun"},
		{aspect(zodiac.Mercury, sky.Conjunction), Neutral, "mixed nature"},
	}
	for _, tt := range tests {
		js := Timing(testSky(zodiac.Sun, zodiac.Sun, tt.next), "", nil)
		if j := js[0]; j.About != "Moon" || j.Effect != tt.effect || !strings.Contains(j.Text, tt.text) {
			t.Errorf("next %+v: %+v, want %v containing %q", tt.next, j, tt.effect, tt.text)
		}
	}

	// The next aspect to a significator of the matter says so.
	js := Timing(testSky(zodiac.Sun, zodiac.Sun, aspect(zodiac.Venus, sky.Sextile)), "marriage", []zodiac.Planet{zodiac.Venus})
	if _, ok := moonJudgement(js, "good sign for the matter. Venus signifies marriage."); !ok {
		t.Errorf("no mention of the significator: %+v", js)
	}

	// Waxing, waning, dignity, and the Sun.
	s := testSky(zodiac.Sun, zodiac.Sun, nil)
	if j, ok := moonJudgement(Timing(s, "", nil), "waxing (crescent, 25% lit)"); !ok || j.Effect != Helps {
		t.Errorf("waxing: %+v", j)
	}
	s.Moon.Phase = sky.Balsamic
	moon := &s.Bodies[zodiac.Moon]
	moon.Sign, moon.Dignities, moon.Solar = zodiac.Scorpio, []sky.Dignity{sky.Fall}, sky.Combust
	js = Timing(s, "", nil)
	for _, w := range []struct {
		text   string
		effect Effect
	}{
		{"waning (balsamic, 25% lit)", Mixed},
		{"The Moon is weak in Scorpio, its fall.", Hinders},
		{"The Moon is combust", Hinders},
	} {
		if j, ok := moonJudgement(js, w.text); !ok || j.Effect != w.effect {
			t.Errorf("%q: %+v (found %v)", w.text, j, ok)
		}
	}
	moon.Sign, moon.Dignities, moon.Solar = zodiac.Taurus, []sky.Dignity{sky.Exaltation}, sky.NotNearSun
	if j, ok := moonJudgement(Timing(s, "", nil), "The Moon is strong in Taurus, its exaltation."); !ok || j.Effect != Helps {
		t.Errorf("exalted: %+v", j)
	}
}
