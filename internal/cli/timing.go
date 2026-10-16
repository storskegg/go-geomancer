package cli

import (
	"unicode"
	"unicode/utf8"

	"github.com/storskegg/go-geomancer/chart"
	"github.com/storskegg/go-geomancer/content"
	"github.com/storskegg/go-geomancer/zodiac"
)

// timing judges the moment of the reading against its matter, or returns
// nil if the moment isn't known.
func timing(r chart.Reading, c *content.Set) []chart.Judgement {
	if r.Sky == nil {
		return nil
	}
	subject, natural := matterPlanets(r, c)
	return chart.Timing(r.Sky, subject, natural)
}

// matterPlanets names the reading's matter and its natural significators: the
// question's topics within the significator's house (or, for a derived
// house, within the subject's own), or failing those, the house's principal
// topic. It returns nothing for a reading without a significator.
func matterPlanets(r chart.Reading, c *content.Set) (string, []zodiac.Planet) {
	house, topics := r.Significator, []string(nil)
	if m := r.Matter; m != nil {
		topics = m.Topics
		if m.Derived != nil {
			house = m.Derived.Subject
		}
	}
	sig, ok := c.Signification(house)
	if !ok {
		return "", nil
	}
	subject := lowerFirst(r.Matter.Subject())
	if subject != "" && r.Matter.Derived != nil {
		subject = "the " + subject // "the husband's career"
	}
	if subject == "" && len(sig.Topics) > 0 {
		subject = sig.Topics[0].Name
	}
	return subject, sig.Planets(topics)
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}
