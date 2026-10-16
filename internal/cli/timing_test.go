package cli

import (
	"slices"
	"testing"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/zodiac"
)

func TestMatterPlanets(t *testing.T) {
	husband := &chart.Derivation{Person: 7, Subject: 10, PersonWord: "husband", SubjectWord: "job"}
	tests := []struct {
		name    string
		sig     int
		matter  *chart.Matter
		subject string
		planets []zodiac.Planet
	}{
		{"topic", 6, &chart.Matter{Topics: []string{"small animals"}}, "small animals", []zodiac.Planet{zodiac.Mercury}},
		{"two topics", 6, &chart.Matter{Topics: []string{"sickness", "small animals"}}, "sickness and small animals", []zodiac.Planet{zodiac.Mars, zodiac.Saturn, zodiac.Mercury}},
		{"unknown: the principal topic", 7, nil, "marriage", []zodiac.Planet{zodiac.Venus}},
		// The subject's topics are in its own house, the 10th.
		{"derived", 4, &chart.Matter{Topics: []string{"career"}, Derived: husband}, "the husband's career", []zodiac.Planet{zodiac.Sun}},
		{"no significator", 0, nil, "", nil},
	}
	for _, tt := range tests {
		r := chart.Reading{Significator: tt.sig, Matter: tt.matter}
		subject, planets := matterPlanets(r, content.Default())
		if subject != tt.subject || !slices.Equal(planets, tt.planets) {
			t.Errorf("%s: %q %v, want %q %v", tt.name, subject, planets, tt.subject, tt.planets)
		}
	}
}
