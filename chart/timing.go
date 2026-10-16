package chart

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/storskegg/go-geomancer/sky"
	"github.com/storskegg/go-geomancer/zodiac"
)

// Effect is how a point of timing bears on the matter.
type Effect uint8

const (
	Helps Effect = iota
	Hinders
	Mixed
	Neutral
)

var effectLabels = []string{"helps", "hinders", "mixed", "neutral"}

func (e Effect) String() string {
	if int(e) < len(effectLabels) {
		return effectLabels[e]
	}
	return fmt.Sprintf("Effect(%d)", e)
}

func (e Effect) MarshalText() ([]byte, error) {
	if int(e) >= len(effectLabels) {
		return nil, fmt.Errorf("chart: unknown effect %d", e)
	}
	return []byte(effectLabels[e]), nil
}

func (e *Effect) UnmarshalText(b []byte) error {
	i := slices.Index(effectLabels, string(b))
	if i < 0 {
		return fmt.Errorf("chart: unknown effect %q", b)
	}
	*e = Effect(i)
	return nil
}

// Judgement is one way the moment of the reading bears on its matter.
type Judgement struct {
	// About is what's judged: "Hour", "Day" or "Moon".
	About  string `json:"about"`
	Effect Effect `json:"effect"`
	Text   string `json:"text"`
}

// Timing judges the moment of a reading against its matter, in the
// traditional manner: whether the planetary hour and day are of a planet
// that signifies the matter (natural, such as Venus for marriage, named by
// subject), or of a benefic or malefic; and the Moon's condition, which
// bears on every question: void of course, its next aspect, waxing or
// waning, its dignity and its closeness to the Sun. With no subject, as for
// a reading without a significator, only the Moon is judged.
func Timing(s *sky.Sky, subject string, natural []zodiac.Planet) []Judgement {
	var out []Judgement
	if subject != "" && len(natural) > 0 {
		if h := s.Day.Hour; h != nil {
			out = append(out, rulerJudgement("Hour", "hour", h.Ruler, subject, natural))
		} else {
			out = append(out, Judgement{"Hour", Neutral, "No planetary hour: " + s.Day.Unavailable + "."})
		}
		out = append(out, rulerJudgement("Day", "day", s.Day.Ruler, subject, natural))
	}
	return append(out, moonJudgements(s, subject, natural)...)
}

// nature is a planet's traditional nature.
type nature uint8

const (
	common nature = iota // the Sun, Moon and Mercury, as their company makes them
	benefic
	malefic
)

func natureOf(p zodiac.Planet) nature {
	switch p {
	case zodiac.Jupiter, zodiac.Venus:
		return benefic
	case zodiac.Mars, zodiac.Saturn:
		return malefic
	case zodiac.Sun, zodiac.Moon, zodiac.Mercury, zodiac.NorthNode, zodiac.SouthNode:
	}
	return common
}

// epithets name the benefics and malefics.
var epithets = map[zodiac.Planet]string{
	zodiac.Jupiter: "the greater benefic", zodiac.Venus: "the lesser benefic",
	zodiac.Saturn: "the greater malefic", zodiac.Mars: "the lesser malefic",
}

// the names a planet mid-sentence: "Venus", but "the Sun".
func the(p zodiac.Planet) string {
	if p == zodiac.Sun || p == zodiac.Moon {
		return "the " + p.String()
	}
	return p.String()
}

// list joins planets' names: "Venus", "Mars and Saturn".
func list(ps []zodiac.Planet) string {
	names := make([]string, len(ps))
	for i, p := range ps {
		names[i] = the(p)
	}
	if len(names) <= 1 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}

// signify writes "Venus signifies" or "Mars and Saturn signify".
func signify(ps []zodiac.Planet) string {
	if len(ps) == 1 {
		return list(ps) + " signifies"
	}
	return list(ps) + " signify"
}

// rulerJudgement judges an hour or day ruled by p.
func rulerJudgement(about, period string, p zodiac.Planet, subject string, natural []zodiac.Planet) Judgement {
	lead := "The " + period + " of " + the(p)
	switch {
	case slices.Contains(natural, p):
		return Judgement{about, Helps, fmt.Sprintf("%s, natural significator of %s, suits the question.", lead, subject)}
	case natureOf(p) == benefic:
		return Judgement{about, Helps, fmt.Sprintf("%s, %s, favours the question, though %s %s.", lead, epithets[p], signify(natural), subject)}
	case natureOf(p) == malefic:
		return Judgement{about, Hinders, fmt.Sprintf("%s, %s, is unfortunate for %s, which %s.", lead, epithets[p], subject, signify(natural))}
	}
	return Judgement{about, Neutral, fmt.Sprintf("%s neither suits nor hinders %s, which %s.", lead, subject, signify(natural))}
}

// when writes a time with its weekday: "3:57 pm Friday".
func when(t time.Time) string { return t.Format("3:04 pm Monday") }

func moonJudgements(s *sky.Sky, subject string, natural []zodiac.Planet) []Judgement {
	m := s.Moon
	moon := s.Body(zodiac.Moon)
	var out []Judgement

	if m.VoidOfCourse {
		out = append(out, Judgement{"Moon", Hinders, fmt.Sprintf(
			"The Moon is void of course: it makes no aspect before leaving %s at %s. Traditionally nothing comes of the matter, or not as hoped.",
			moon.Sign, when(m.LeavesSign))})
	} else if a := m.NextAspect; a != nil {
		out = append(out, nextAspect(a, subject, natural))
	}

	phase := fmt.Sprintf("%s, %.0f%% lit", strings.ToLower(m.Phase.String()), m.Illuminated*100)
	if m.Phase.Waxing() {
		out = append(out, Judgement{"Moon", Helps, "The Moon is waxing (" + phase + "): a time of increase, favouring growth and new undertakings."})
	} else {
		out = append(out, Judgement{"Moon", Mixed, "The Moon is waning (" + phase + "): a time of decrease, favouring endings and letting go rather than beginnings."})
	}

	switch {
	case slices.Contains(moon.Dignities, sky.Domicile):
		out = append(out, Judgement{"Moon", Helps, fmt.Sprintf("The Moon is strong in %s, its own sign.", moon.Sign)})
	case slices.Contains(moon.Dignities, sky.Exaltation):
		out = append(out, Judgement{"Moon", Helps, fmt.Sprintf("The Moon is strong in %s, its exaltation.", moon.Sign)})
	case slices.Contains(moon.Dignities, sky.Detriment):
		out = append(out, Judgement{"Moon", Hinders, fmt.Sprintf("The Moon is weak in %s, its detriment.", moon.Sign)})
	case slices.Contains(moon.Dignities, sky.Fall):
		out = append(out, Judgement{"Moon", Hinders, fmt.Sprintf("The Moon is weak in %s, its fall.", moon.Sign)})
	}

	switch moon.Solar {
	case sky.Cazimi:
		out = append(out, Judgement{"Moon", Helps, "The Moon is cazimi, in the heart of the Sun: greatly strengthened."})
	case sky.Combust:
		out = append(out, Judgement{"Moon", Hinders, "The Moon is combust, within 8°30′ of the Sun: weakened, and the matter obscure."})
	case sky.UnderTheBeams:
		out = append(out, Judgement{"Moon", Hinders, "The Moon is under the Sun's beams: weakened, and the matter not yet clear."})
	case sky.NotNearSun:
	}
	return out
}

// nextAspect judges the Moon's next aspect by its kind and the planet's
// nature. Easy aspects are the sextile and trine; hard, the square and
// opposition.
func nextAspect(a *sky.MoonAspect, subject string, natural []zodiac.Planet) Judgement {
	easy := a.Kind == sky.Sextile || a.Kind == sky.Trine
	hard := a.Kind == sky.Square || a.Kind == sky.Opposition
	effect, why := Neutral, "a meeting with a planet of mixed nature"
	switch n := natureOf(a.Planet); {
	case n == benefic && (easy || !hard):
		effect, why = Helps, "a benefic, well met: a good sign for the matter"
	case n == benefic:
		effect, why = Mixed, "a benefic, though by a hard aspect: help, with difficulty"
	case n == malefic && easy:
		effect, why = Mixed, "a malefic, though by an easy aspect: difficulty, eased"
	case n == malefic:
		effect, why = Hinders, "a malefic, badly met: obstacles to the matter"
	case easy:
		effect, why = Helps, "an easy aspect: things go smoothly"
	case hard:
		effect, why = Hinders, "a hard aspect: things meet resistance"
	}
	why = capitalize(why) + "."
	if subject != "" && slices.Contains(natural, a.Planet) {
		why += fmt.Sprintf(" %s signifies %s.", capitalize(the(a.Planet)), subject)
	}
	kind, to := "a "+strings.ToLower(a.Kind.String()), "to"
	switch a.Kind {
	case sky.Opposition:
		kind = "an opposition"
	case sky.Conjunction:
		to = "with"
	case sky.Sextile, sky.Square, sky.Trine:
	}
	return Judgement{"Moon", effect, fmt.Sprintf("The Moon's next aspect is %s %s %s, exact at %s. %s",
		kind, to, the(a.Planet), when(a.Exact), why)}
}
