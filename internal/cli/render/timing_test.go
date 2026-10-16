package render

import (
	"strings"
	"testing"

	"github.com/storskegg/go-geomancer/chart"
)

func TestTiming(t *testing.T) {
	js := []chart.Judgement{
		{About: "Hour", Effect: chart.Helps, Text: "The hour of Venus suits the question."},
		{About: "Day", Effect: chart.Hinders, Text: "The day of Saturn doesn't."},
		{About: "Moon", Effect: chart.Mixed, Text: "The Moon is waning."},
		{About: "Moon", Effect: chart.Neutral, Text: "Neither."},
	}
	out := plain(Timing(js, Options{}))
	for _, want := range []string{"✓ Hour  The hour of Venus", "✗ Day   The day of Saturn", "± Moon  The Moon is waning.", "· Moon  Neither."} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	ascii := plain(Timing(js, Options{Symbols: ASCII}))
	for _, want := range []string{"+ Hour", "- Day ", "~ Moon", ". Moon"} {
		if !strings.Contains(ascii, want) {
			t.Errorf("ASCII output lacks %q:\n%s", want, ascii)
		}
	}
}
